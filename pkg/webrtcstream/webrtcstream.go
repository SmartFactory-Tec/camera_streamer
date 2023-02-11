package webrtcstream

import (
	"camera_server/pkg/gst"
	"context"
	"fmt"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"sync"
)

type Config struct {
	Name     string `toml:"name"`
	Id       string `toml:"id"`
	Hostname string `toml:"hostname"`
	Path     string `toml:"path"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type WebRTCStream struct {
	Id   string `json:"id"`
	Name string `json:"name"`

	pipeline          *gst.Pipeline
	bus               *gst.Bus
	busMessageHandler BusMessageHandlerFunc

	source *gst.UriDecodeBin

	// Shared elements across all tracks
	queue      *gst.Queue
	enc        *gst.Vp8Enc
	sourceTee  *gst.Tee
	multiqueue *gst.Multiqueue

	// Pads between source and multiqueue and multiqueue and sinks
	sourceTeeSrcPads   map[int]*gst.Pad
	multiqueueSinkPads map[int]*gst.Pad
	multiqueueSrcPads  map[int]*gst.Pad

	// Maps for per track elements
	sinks map[int]*WebRtcSink // Map of track ID to appsink elements

	streamMu    sync.Mutex
	sinkCounter int
}

// New constructs a stream with a given name and id that pulls data from a given source gst element.
// The element is expected to provide the stream with x-raw video, so it must decode any video it sends.
func New(config Config) (*WebRTCStream, error) {
	var err error
	defer func() {
		err = fmt.Errorf("could not create stream: %w", err)
	}()

	// First create the pipeline
	pipeline, err := gst.NewGstPipeline(fmt.Sprintf("%s-pipeline", config.Id))
	if err != nil {
		return nil, err
	}

	bus, err := pipeline.Bus()
	if err != nil {
		return nil, err
	}

	var uri string

	// Create source uridecodebin
	if config.User != "" && config.Password != "" {
		uri = fmt.Sprintf("rtsp://%s:%s@%s:%d%s", config.User, config.Password, config.Hostname, config.Port, config.Path)
	} else {
		uri = fmt.Sprintf("rtsp://%s:%d%s", config.Hostname, config.Port, config.Path)
	}

	src, err := gst.NewUriDecodeBin(config.Id, uri)

	// Create shared elements (tee, encoder and multiqueue)
	queue, err := gst.NewQueue(fmt.Sprintf("%s-queue", config.Id))
	if err != nil {
		return nil, err
	}

	enc, err := gst.NewVp8Enc(fmt.Sprintf("%s-enc", config.Id))
	if err != nil {
		return nil, err
	}
	// realtime
	enc.SetProperty("deadline", 33333)
	enc.SetProperty("cpu-used", 3)

	srcTee, err := gst.NewTee(fmt.Sprintf("%s-source-tee", config.Id))
	if err != nil {
		return nil, err
	}
	srcTee.SetProperty("allow-not-linked", true)

	multiqueue, err := gst.NewMultiqueue(fmt.Sprintf("%s-multiqueue", config.Id))
	if err != nil {
		return nil, err
	}

	// Build pipeline
	pipeline.AddElement(src)
	pipeline.AddElement(queue)
	pipeline.AddElement(enc)
	pipeline.AddElement(srcTee)
	pipeline.AddElement(multiqueue)

	gst.LinkElements(queue, enc)
	gst.LinkElements(enc, srcTee)

	src.OnPadAdded(func(pad *gst.Pad) {
		// TODO handle error
		caps, err := pad.Caps()
		format, err := caps.Format(0)
		if err != nil {
			panic(err)
		}

		if format.Name() != "video/x-raw" {
			return
		}

		sinkPad, ok := queue.GetPad("sink")

		if !ok {
			panic("failed getting sink pad of queue")
		}

		err = gst.LinkPads(pad, sinkPad)
		if err != nil {
			panic(err)
		}
	})

	stream := WebRTCStream{
		config.Id,
		config.Name,
		pipeline,
		bus,
		nil,
		src,
		queue,
		enc,
		srcTee,
		multiqueue,
		make(map[int]*gst.Pad),
		make(map[int]*gst.Pad),
		make(map[int]*gst.Pad),
		make(map[int]*WebRtcSink),
		sync.Mutex{},
		0,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go stream.processMsgBus(ctx)

	// Make sure that if this object is ever garbage collected, cancel the bus context
	runtime.SetFinalizer(&stream, func(stream *WebRTCStream) {
		cancel()
	})

	return &stream, nil
}

// removeTrack stops execution of a give webrtc track by removing its sink and stopping the pipeline if it's the last
// track playing.
func (s *WebRTCStream) removeTrack(track *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) error {
	logger = logger.Named("removeTrack").With("id", track.StreamID())
	s.streamMu.Lock()
	defer s.streamMu.Unlock()

	logger.Debugw("removing track")

	if len(s.sinks) == 0 {
		return fmt.Errorf("no tracks currently playing")
	}

	parsedInt, err := strconv.ParseInt(track.StreamID(), 10, 0)
	trackId := int(parsedInt)

	if err != nil {
		return fmt.Errorf("invalid track stream id: %w", err)
	}

	sink := s.sinks[trackId]
	sourceTeePad := s.sourceTeeSrcPads[trackId]
	multiqueueSinkPad := s.multiqueueSinkPads[trackId]

	delete(s.sinks, trackId)
	delete(s.sourceTeeSrcPads, trackId)
	delete(s.multiqueueSrcPads, trackId)
	delete(s.multiqueueSinkPads, trackId)

	if len(s.sinks) == 0 {
		logger.Debugw("no tracks left, pausing pipeline")
		if err := s.pipeline.SetState(gst.READY); err != nil {
			return err
		}
	}

	if err := sink.SetState(gst.NULL); err != nil {
		return err
	}

	s.pipeline.RemoveElement(sink)
	s.sourceTee.ReleaseRequestPad(sourceTeePad)
	s.multiqueue.ReleaseRequestPad(multiqueueSinkPad)

	return nil
}

// GenerateTrack generates a new track with a given resolution from the stream source
// It creates new elements as needed, reusing them if they already exist.
// (Hopefully) Concurrency safe
func (s *WebRTCStream) createTrack(ctx context.Context, logger *zap.SugaredLogger) (track *webrtc.TrackLocalStaticSample, err error) {
	logger = logger.Named("createTrack")
	defer func() {
		if err != nil {
			// Free pads if track creation failed
			if pad, ok := s.sourceTeeSrcPads[s.sinkCounter]; ok {
				s.sourceTee.ReleaseRequestPad(pad)
				delete(s.sourceTeeSrcPads, s.sinkCounter)
			}
			if pad, ok := s.multiqueueSrcPads[s.sinkCounter]; ok {
				s.multiqueue.ReleaseRequestPad(pad)
				delete(s.multiqueueSrcPads, s.sinkCounter)
			}
			if _, ok := s.multiqueueSinkPads[s.sinkCounter]; ok {
				delete(s.multiqueueSinkPads, s.sinkCounter)
			}

		}
	}()

	s.streamMu.Lock()
	defer s.streamMu.Unlock()

	logger.Debugw("creating track", "id", s.sinkCounter)
	// first create the webrtc track and the sink
	track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: "video/vp8",
	}, "video", strconv.Itoa(s.sinkCounter))
	if err != nil {
		return nil, err
	}

	logger.Debugw("creating webrtc sink")
	webrtcSink, err := NewWebRtcSink(fmt.Sprintf("%d-sink", s.sinkCounter), track)
	if err != nil {
		return nil, err
	}

	sourceTeePad, err := s.sourceTee.RequestPad("src_%u")
	if err != nil {
		return nil, err
	}
	s.sourceTeeSrcPads[s.sinkCounter] = sourceTeePad

	mqSinkPad, err := s.multiqueue.RequestPad(fmt.Sprintf("sink_%d", s.sinkCounter))
	if err != nil {
		return nil, err
	}
	s.multiqueueSinkPads[s.sinkCounter] = mqSinkPad

	mqSourcePad, ok := s.multiqueue.GetPad(fmt.Sprintf("src_%d", s.sinkCounter))
	if !ok {
		return nil, fmt.Errorf("could not get multiqueue src pad with id %d", s.sinkCounter)
	}
	s.multiqueueSrcPads[s.sinkCounter] = mqSourcePad

	sinkPad, ok := webrtcSink.GetPad("sink")
	if !ok {
		return nil, err
	}

	s.pipeline.AddElement(webrtcSink)

	err = gst.LinkPads(sourceTeePad, mqSinkPad)
	if err != nil {
		return nil, err
	}

	err = gst.LinkPads(mqSourcePad, sinkPad)
	if err != nil {
		return nil, err
	}

	// If this is the only track
	if len(s.sinks) == 0 {
		logger.Debugw("starting pipeline")
		err = s.pipeline.SetState(gst.PLAYING)
	} else {
		logger.Debugw("joining already running pipeline")
		err = webrtcSink.SetState(gst.PLAYING)
		if err != nil {
			return nil, err
		}
	}

	s.sinks[s.sinkCounter] = webrtcSink

	s.sinkCounter++

	go webrtcSink.Start(ctx)

	return
}

type TrackRequestHandler func(ctx context.Context, track webrtc.TrackLocal)

func (s *WebRTCStream) HandleTrackRequest(ctx context.Context, logger *zap.SugaredLogger, handler TrackRequestHandler) {
	logger = logger.Named("HandleTrackRequest").With("stream id", s.Id)

	ctx, cancelTrack := context.WithCancel(ctx)

	logger.Debugw("creating track from stream")

	track, err := s.createTrack(ctx, logger)
	if err != nil {
		logger.Error(fmt.Errorf("error creating track: %w", err))
		cancelTrack()
		return
	}

	handler(ctx, track)

	cancelTrack()

	logger.Debugw("removing track from stream")
	err = s.removeTrack(track, logger)
	if err != nil {
		logger.Error(fmt.Errorf("error cleaning up track: %w", err))
		return
	}
}

type BusMessageHandlerFunc func(ctx context.Context, message *gst.Message)

func (s *WebRTCStream) OnBusMessage(handler BusMessageHandlerFunc) {
	s.busMessageHandler = handler
}

func (s *WebRTCStream) processMsgBus(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := s.bus.PopMessageWithFilter(gst.ERROR | gst.END_OF_STREAM)
			// If there's an error, there's no message to process
			if err == nil && s.busMessageHandler != nil {
				s.busMessageHandler(ctx, msg)
			}
		}

	}
}

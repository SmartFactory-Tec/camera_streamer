package webrtcstream

import (
	"context"
	"fmt"
	"github.com/SmartFactory-Tec/camera_server/pkg/gst"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Name             string `toml:"name" json:"name"`
	Id               int    `toml:"id" json:"id"`
	ConnectionString string `toml:"connection_string" json:"connection_string"`
}

type WebRTCStream struct {
	Id   int    `json:"id"`
	Name string `json:"name"`

	pipeline *gst.Pipeline
	bus      *gst.Bus

	source       *gst.RtspSource
	sourceLinked bool

	// Shared elements across all tracks
	queue      *gst.Queue
	dec        *gst.DecodeBin3
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

	// shared elements
	src, err := gst.NewRtspSource(fmt.Sprintf("%s-src", config.Id), config.ConnectionString)
	if err != nil {
		return nil, err
	}
	src.SetProperty("latency", 200)   // in ms
	src.SetProperty("buffer-mode", 3) // slave to sender, the camera
	src.SetProperty("ntp-sync", true)

	dec, err := gst.NewDecodeBin3(fmt.Sprintf("%s-dec", config.Id))
	if err != nil {
		return nil, err
	}

	queue, err := gst.NewQueue(fmt.Sprintf("%s-queue", config.Id))
	if err != nil {
		return nil, err
	}
	queue.SetProperty("leaky", 2)
	queue.SetProperty("max-size-buffers", 1)

	enc, err := gst.NewVp8Enc(fmt.Sprintf("%s-enc", config.Id))
	if err != nil {
		return nil, err
	}
	// realtime
	//enc.SetProperty("deadline", 1)
	enc.SetProperty("deadline", 30000)
	enc.SetProperty("cpu-used", 0)
	enc.SetProperty("bits-per-pixel", float32(0.04))
	enc.SetProperty("end-usage", 1)
	//enc.SetProperty("undershoot", 95)
	enc.SetProperty("error-resilient", 0x1)

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
	pipeline.AddElement(dec)
	pipeline.AddElement(enc)
	pipeline.AddElement(srcTee)
	pipeline.AddElement(multiqueue)

	gst.LinkElements(queue, enc)
	gst.LinkElements(enc, srcTee)

	stream := WebRTCStream{
		config.Id,
		config.Name,
		pipeline,
		bus,
		src,
		false,
		queue,
		dec,
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

	src.OnPadAdded(func(pad *gst.Pad) {
		if stream.sourceLinked {
			return
		}
		sinkPad, ok := dec.GetPad("sink")

		if !ok {
			panic("failed getting sink pad of queue")
		}

		err = gst.LinkPads(pad, sinkPad)
		if err != nil {
			panic(err)
		}

		stream.sourceLinked = true
	})

	dec.OnPadAdded(func(pad *gst.Pad) {
		if pad.Name() != "video_0" {
			return
		}

		sinkPad, ok := queue.GetPad("sink")
		if !ok {
			panic("failed getting sink pad of encoder")
		}

		err = gst.LinkPads(pad, sinkPad)
		if err != nil {
			panic(err)
		}
	})

	err = stream.pipeline.SetState(gst.PAUSED)
	if err != nil {
		return nil, fmt.Errorf("error initializing the pipeline: %w", err)
	}

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
		if err := s.pipeline.SetState(gst.PAUSED); err != nil {
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

	go s.processMsgBus(ctx, logger)

	logger.Debugw("creating track from stream")

	track, err := s.createTrack(ctx, logger)
	if err != nil {
		logger.Error(fmt.Errorf("error creating track: %w", err))
		cancelTrack()
		return
	}

	min, max, err := s.pipeline.QueryLatency()
	if err != nil {
		logger.Errorw("could not query latency")
	}

	logger.Debugw("latency", "min", min, "max", max)

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

func (s *WebRTCStream) processMsgBus(ctx context.Context, logger *zap.SugaredLogger) {
	logger = logger.Named("MsgBus")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := s.bus.PopMessageWithFilter(gst.LATENCY)

			// If there's an error, there's no message to process
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			switch msg.Type() {
			case gst.LATENCY:
				//latency := s.pipeline.Latency()
				//if latency == -1 {
				//	s.pipeline.SetProperty("latency", 0)
				//} else {
				//	logger.Debugw("track latency", "latency", s.pipeline.Latency().Milliseconds())
				//}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}

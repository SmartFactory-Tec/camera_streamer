package webrtcstream

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SmartFactory-Tec/camera_streamer/pkg/gst"
	"github.com/hashicorp/go-multierror"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type Orientation string

const (
	CameraOrientationVertical           Orientation = "vertical"
	CameraOrientationHorizontal         Orientation = "horizontal"
	CameraOrientationInvertedVertical   Orientation = "inverted_vertical"
	CameraOrientationInvertedHorizontal Orientation = "inverted_horizontal"
)

func (co *Orientation) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*co))
}

func (co *Orientation) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	if s == string(CameraOrientationHorizontal) ||
		s == string(CameraOrientationVertical) ||
		s == string(CameraOrientationInvertedHorizontal) ||
		s == string(CameraOrientationInvertedVertical) {
		*co = Orientation(s)
		return nil
	} else {
		return fmt.Errorf("invalid value for enum CameraOrientation")
	}
}

type Config struct {
	Name             string      `toml:"name" json:"name"`
	Id               int         `toml:"id" json:"id"`
	ConnectionString string      `toml:"connection_string" json:"connection_string"`
	Orientation      Orientation `toml:"orientation" json:"orientation"`
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

func orientationToMethod(orientation Orientation) gst.VideoOrientationMethod {
	switch orientation {
	case CameraOrientationHorizontal:
		return gst.IDENTITY
	case CameraOrientationInvertedHorizontal:
		return gst.ROTATE_180
	case CameraOrientationVertical:
		return gst.ROTATE_CLOCKWISE_90
	case CameraOrientationInvertedVertical:
		return gst.ROTATE_COUNTERCLOCKWISE_90
	}
	return gst.IDENTITY
}

// New constructs a stream with a given name and id that pulls data from a given source gst element.
// The element is expected to provide the stream with x-raw video, so it must decode any video it sends.
func New(config Config) (*WebRTCStream, error) {
	// First create the pipeline
	pipeline, err := gst.NewGstPipeline(fmt.Sprintf("%d-pipeline", config.Id))
	if err != nil {
		return nil, err
	}

	bus, err := pipeline.Bus()
	if err != nil {
		return nil, err
	}

	// build pipeline elements
	var result *multierror.Error
	src, err := gst.NewRtspSource(fmt.Sprintf("%d-src", config.Id), config.ConnectionString)
	result = multierror.Append(result, err)
	queue, err := gst.NewQueue(fmt.Sprintf("%d-queue", config.Id))
	result = multierror.Append(result, err)
	videoFlip, err := gst.NewVideoFlip(fmt.Sprintf("%d-videoflip", config.Id), orientationToMethod(config.Orientation))
	result = multierror.Append(result, err)
	dec, err := gst.NewDecodeBin3(fmt.Sprintf("%d-dec", config.Id))
	result = multierror.Append(result, err)
	enc, err := gst.NewVp8Enc(fmt.Sprintf("%d-enc", config.Id))
	result = multierror.Append(result, err)
	srcTee, err := gst.NewTee(fmt.Sprintf("%d-source-tee", config.Id))
	result = multierror.Append(result, err)
	multiqueue, err := gst.NewMultiqueue(fmt.Sprintf("%d-multiqueue", config.Id))
	result = multierror.Append(result, err)

	if result.ErrorOrNil() != nil {
		return nil, result
	}

	// configure source
	result = nil
	result = multierror.Append(result, src.SetProperty("latency", 200))   // in ms
	result = multierror.Append(result, src.SetProperty("buffer-mode", 3)) // slave to sender, the camera
	result = multierror.Append(result, src.SetProperty("ntp-sync", true))
	if result.ErrorOrNil() != nil {
		return nil, result
	}

	//configure queue
	result = nil
	//result = multierror.Append(result, queue.SetProperty("leaky", 2))
	result = multierror.Append(result, queue.SetProperty("max-size-buffers", 1))
	if result.ErrorOrNil() != nil {
		return nil, result
	}

	//configure encoder
	result = nil
	result = multierror.Append(result, enc.SetProperty("deadline", 30000))
	result = multierror.Append(result, enc.SetProperty("cpu-used", 4))
	result = multierror.Append(result, enc.SetProperty("bits-per-pixel", float32(0.02)))
	result = multierror.Append(result, enc.SetProperty("end-usage", 0))
	result = multierror.Append(result, enc.SetProperty("error-resilient", 0x1))
	if result.ErrorOrNil() != nil {
		return nil, result
	}

	// configure tee
	err = srcTee.SetProperty("allow-not-linked", true)
	if err != nil {
		return nil, err
	}

	// Build pipeline
	pipeline.AddElement(src)
	pipeline.AddElement(queue)
	pipeline.AddElement(videoFlip)
	pipeline.AddElement(dec)
	pipeline.AddElement(enc)
	pipeline.AddElement(srcTee)
	pipeline.AddElement(multiqueue)

	// Link pipeline together
	result = nil
	result = multierror.Append(result, gst.LinkElements(queue, videoFlip))
	result = multierror.Append(result, gst.LinkElements(videoFlip, enc))
	result = multierror.Append(result, gst.LinkElements(enc, srcTee))
	if result.ErrorOrNil() != nil {
		return nil, err
	}

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

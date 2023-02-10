package main

import (
	"camera_server/pkg/gst"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"sync"
)

type Stream struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	pipeline     *gst.Pipeline
	sourceTee    *gst.Tee
	queues       *sync.Map
	sinks        *sync.Map
	trackCtxs    *sync.Map
	trackCancels *sync.Map
	pipelineLock *sync.Mutex
	trackCount   int
}

// NewStream constructs a stream with a given name and id that pulls data from a given source gst element.
// The element is expected to provide the stream with x-raw video, so it must decode any video it sends.
func NewStream(name string, id string, src *gst.UriDecodeBin) (Stream, error) {
	// Create the stream pipeline
	newPipeline, err := gst.NewGstPipeline(id)

	if err != nil {
		return Stream{}, err
	}

	srcTee, err := gst.NewTee(fmt.Sprintf("%s-source-tee", id))
	if err != nil {
		return Stream{}, err
	}

	srcTee.SetProperty("allow-not-linked", true)

	enc, err := gst.NewVp8Enc(fmt.Sprintf("%s-enc", id))
	if err != nil {
		return Stream{}, err
	}
	enc.SetProperty("deadline", 1)

	srcQueue, err := gst.NewQueue(fmt.Sprintf("%s-srcQueue", id))
	if err != nil {
		return Stream{}, err
	}

	// Build pipeline
	newPipeline.AddElement(src)
	newPipeline.AddElement(srcQueue)
	newPipeline.AddElement(enc)
	newPipeline.AddElement(srcTee)

	gst.LinkElements(srcQueue, enc)
	gst.LinkElements(enc, srcTee)

	src.OnPadAdded(func(pad *gst.Pad) {
		// TODO handle error
		caps, err := pad.Caps()
		format, err := caps.Format(0)
		panicIfError(err)

		if format.Name() != "video/x-raw" {
			return
		}

		sinkPad, err := srcQueue.QueryPadByName("sink")
		err = gst.LinkPads(pad, sinkPad)
		if err != nil {
			panic(err)
		}
	})

	stream := Stream{
		id,
		name,
		newPipeline,
		srcTee,
		new(sync.Map),
		new(sync.Map),
		new(sync.Map),
		new(sync.Map),
		new(sync.Mutex),
		0,
	}

	go stream.MsgBus()

	return stream, nil
}

// StopTrack stops execution of a give webrtc track by removing its sink and stopping the pipeline if it's the last
// track playing.
func (s *Stream) StopTrack(track *webrtc.TrackLocalStaticSample) error {
	s.pipelineLock.Lock()
	defer s.pipelineLock.Unlock()

	if s.trackCount >= 1 {
		s.trackCount--
	}

	if s.trackCount == 0 {
		if err := s.pipeline.SetState(gst.READY); err != nil {
			return err
		}
	}

	trackID := track.StreamID()

	var cancel context.CancelFunc
	if value, ok := s.trackCancels.LoadAndDelete(trackID); ok {
		cancel = value.(context.CancelFunc)
	} else {
		return fmt.Errorf("unable to retrieve end channel for track")
	}

	cancel()

	var sink WebRtcSink
	if value, ok := s.sinks.LoadAndDelete(trackID); ok {
		sink = value.(WebRtcSink)
	} else {
		return fmt.Errorf("unable to retrieve webrtc sink for track")
	}

	var queue *gst.Queue2
	if value, ok := s.queues.LoadAndDelete(trackID); ok {
		queue = value.(*gst.Queue2)
	} else {
		return fmt.Errorf("Unable to retrieve webrtc sink for track")
	}

	if err := sink.SetState(gst.NULL); err != nil {
		return err
	}

	if err := queue.SetState(gst.NULL); err != nil {
		return err
	}

	s.pipeline.RemoveElement(sink)
	s.pipeline.RemoveElement(queue)

	return nil
}

// GenerateTrack generates a new track with a given resolution from the stream source
// It creates new elements as needed, reusing them if they already exist.
// (Hopefully) Concurrency safe
func (s *Stream) GenerateTrack() (*webrtc.TrackLocalStaticSample, error) {
	s.pipelineLock.Lock()
	defer s.pipelineLock.Unlock()

	trackID := fmt.Sprintf("%s-stream", uuid.New())

	queue, err := gst.NewQueue2(fmt.Sprintf("%s-queue", trackID))
	if err != nil {
		return nil, err
	}
	s.queues.Store(trackID, queue)

	// this implementation uses the VP8 codec, without any option of dynamically changing it.
	// Chosen because it is the most widely supported webrtc codec
	track, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: "video/vp8",
	}, "video", trackID)
	if err != nil {
		return nil, err
	}

	webrtcSink, err := NewWebRtcSink(fmt.Sprintf("%s-sink", trackID), track)
	if err != nil {
		return nil, err
	}

	s.sinks.Store(trackID, webrtcSink)

	s.pipeline.AddElement(queue)
	s.pipeline.AddElement(webrtcSink)

	if s.trackCount == 0 {

		gst.LinkElements(s.sourceTee, queue)
		gst.LinkElements(queue, webrtcSink)

		if err != nil {
			return nil, err
		}

		err = s.pipeline.SetState(gst.PLAYING)
	} else {
		gst.LinkElements(queue, webrtcSink)

		err := queue.SetState(gst.PLAYING)
		if err != nil {
			return nil, err
		}
		err = webrtcSink.SetState(gst.PLAYING)
		if err != nil {
			return nil, err
		}

		gst.LinkElements(s.sourceTee, queue)
	}

	s.trackCount++

	ctx, cancel := context.WithCancel(context.Background())

	s.trackCtxs.Store(trackID, ctx)
	s.trackCancels.Store(trackID, cancel)

	go webrtcSink.Start(ctx)

	return track, nil
}

func (s *Stream) MsgBus() {
	bus, err := s.pipeline.Bus()
	if err != nil {
		panic(err)
	}

	for {
		msg, err := bus.PopMessageWithFilter(gst.ERROR | gst.END_OF_STREAM)
		// If there's an error, there's no message to process
		if err == nil {
			switch msg.Type() {
			case gst.ERROR:
				_, err := msg.ParseAsError()
				if err != nil {
				}
				return
			case gst.END_OF_STREAM:
				return

			default:
			}
		}
	}
}

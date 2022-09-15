package main

import (
	"camera_server/pkg/gst"
	"camera_server/pkg/gst/elements"
	"fmt"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
)

type CameraStream struct {
	Name     string
	Host     string
	Port     int
	Path     string
	Username string
	Password string

	Pipeline gst.Pipeline

	targetTrackSinks map[string]TrackSink
}

type TrackSink struct {
	queue *elements.Queue
	sink  *WebRtcSink
}

func NewCameraStream(name string, host string, port int, path string, username string, password string, logger *zap.SugaredLogger) CameraStream {
	logger = logger.Named("NewCameraStream").With("camera", name)

	logger.Debugw("Initializing pipeline for camera stream")
	pipeline, err := gst.NewGstPipeline(name)
	panicIfError(err)

	source, err := elements.NewRtspSource(fmt.Sprintf("%s-source", name), fmt.Sprintf("rtsp://%s:%s@%s:%i/%s", username, password, host, port, path))
	panicIfError(err)

	queue, err := elements.NewQueue(fmt.Sprintf("%s-queue", name))
	panicIfError(err)

	depay, err := elements.NewRtpH265Depay(fmt.Sprintf("%s-depay", name))
	panicIfError(err)

	parse, err := elements.NewH265Parse(fmt.Sprintf("%s-parse", name))
	panicIfError(err)

	decode, err := elements.NewAvDecH265(fmt.Sprintf("%s-decode", name))
	panicIfError(err)

	vp8encode, err := elements.NewVp8Enc(fmt.Sprintf("%s-encode", name))
	panicIfError(err)

	tee, err := elements.NewTee(fmt.Sprintf("%s-tee", name))
	panicIfError(err)

	logger.Debugw("Linking elements in pipeline", "pipeline", pipeline.Name())

	if ok := pipeline.AddElement(source); !ok {
		panic("could not add source element")
	}

	if ok := pipeline.AddElement(queue); !ok {
		panic("could not add queue element")
	}

	if ok := pipeline.AddElement(depay); !ok {
		panic("could not add depayloader element")
	}

	if ok := pipeline.AddElement(parse); !ok {
		panic("could not add parser element")
	}

	if ok := pipeline.AddElement(decode); !ok {
		panic("could not add H265 decoder element")
	}

	if ok := pipeline.AddElement(vp8encode); !ok {
		panic("could not add VP8 encoder element")
	}

	if ok := pipeline.AddElement(tee); !ok {
		panic("could not add tee element")
	}

	err = gst.LinkElements(queue, depay)
	panicIfError(err)
	err = gst.LinkElements(depay, parse)
	panicIfError(err)
	err = gst.LinkElements(parse, decode)
	panicIfError(err)
	err = gst.LinkElements(decode, vp8encode)
	panicIfError(err)
	err = gst.LinkElements(vp8encode, tee)
	panicIfError(err)

	source.OnPadAdded(func(newPad gst.Pad) {
		logger := logger.Named("OnPadAdded").With("element", source.Name())
		logger.Debugw("Received new pad")
		format, err := newPad.Format(0)
		panicIfError(err)

		encoding, err := format.QueryStringProperty("encoding-name")
		panicIfError(err)

		if encoding != "H265" {
			logger.Debugw("Ignoring unknown codec", "codec", encoding)
			return
		}

		sinkPad, err := queue.QueryPadByName("sink")
		panicIfError(err)

		logger.Debugw("Linking new source pad")
		err = gst.LinkPads(newPad, &sinkPad)
		panicIfError(err)

	})

	return CameraStream{
		Name:             name,
		Host:             host,
		Port:             port,
		Path:             path,
		Username:         username,
		Password:         password,
		Pipeline:         &pipeline,
		targetTrackSinks: make(map[string]TrackSink),
	}
}

func (c *CameraStream) AddOutputTrack(track *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) {
	logger = logger.Named("AddOutputTrack").With("camera", c.Name, "track", track.ID())
	logger.Debugw("Creating track sink elements")
	queue, err := elements.NewQueue(fmt.Sprintf("%s-%s-queue", c.Name, track.ID()))
	panicIfError(err)
	sink, err := NewWebRtcSink(fmt.Sprintf("%s-%s-sink", c.Name, track.ID()), track)
	panicIfError(err)

	logger.Debugw("Adding track elements to camera stream pipeline", "pipeline", c.Pipeline.Name())
	tee, ok := c.Pipeline.GetElement(fmt.Sprintf("%s-tee", c.Name))
	if !ok {
		logger.Panicw("Could not get tee element for pipeline", "pipeline", c.Pipeline.Name())
	}

	if ok := c.Pipeline.AddElement(queue); !ok {
		logger.Panicw("Could not add queue element to pipeline", "pipeline", c.Pipeline.Name(), "element", queue.Name())
	}
	if ok := c.Pipeline.AddElement(sink); !ok {
		logger.Panicw("Could not add sink element to pipeline", c.Name, "pipeline", c.Pipeline.Name(), "element", sink.Name())
	}

	logger.Debugw("Linking elements to camera stream")
	err = gst.LinkElements(queue, sink)
	panicIfError(err)

	err = gst.LinkElements(tee, queue)
	panicIfError(err)

	if c.Pipeline.State() != gst.PLAYING {
		logger.Debugw("Starting pipeline", "pipeline", c.Pipeline.Name())
		err := c.Pipeline.SetState(gst.PLAYING)
		panicIfError(err)
	} else {
		logger.Debugw("Starting elements in already playing pipeline", "pipeline", c.Pipeline.Name())
		err := queue.SetState(gst.PLAYING)
		panicIfError(err)
		err = sink.SetState(gst.PLAYING)
		panicIfError(err)

	}

	c.targetTrackSinks[track.ID()] = TrackSink{
		&queue, &sink,
	}
}

func (c *CameraStream) RemoveOutputTrack(track *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) error {
	logger = logger.Named("RemoveOutputTrack").With("track", track.ID(), "camera", c.Name)
	trackElements, ok := c.targetTrackSinks[track.ID()]

	if !ok {
		return fmt.Errorf("Track with ID '%s' is not registered with stream '%s'", track.ID(), c.Name)
	}

	if len(c.targetTrackSinks) == 1 {
		logger.Debugw("Stopping pipeline", "pipeline", c.Pipeline.Name())
		err := c.Pipeline.SetState(gst.READY)
		panicIfError(err)
	}

	logger.Debugw("Unloading sink and queue elements", "sink", trackElements.sink.Name(), "queue", trackElements.queue.Name())
	err := trackElements.sink.SetState(gst.NULL)
	panicIfError(err)
	err = trackElements.queue.SetState(gst.NULL)
	panicIfError(err)

	logger.Debugw("Removing sink and queue elements from camera stream pipeline", "pipeline", c.Pipeline.Name(), "sink", trackElements.sink.Name(), "queue", trackElements.queue.Name())
	c.Pipeline.RemoveElement(trackElements.sink)
	c.Pipeline.RemoveElement(trackElements.queue)

	delete(c.targetTrackSinks, track.ID())

	return nil
}

func (c *CameraStream) StartMsgBus(end <-chan bool, logger *zap.SugaredLogger) {
	logger.Debugw("Starting bus listener")
	bus, err := c.Pipeline.Bus()
	panicIfError(err)

	for {
		select {
		case <-end:
			// If received anything through the channel, exit goroutine
			logger.Debugw("Bus listener stopped")
			break
		default:
			// Do nothing
		}

		msg, err := bus.PopMessageWithFilter(gst.ERROR | gst.END_OF_STREAM)
		// If there's an error, there's no message to process
		if err == nil {
			logger.Debugw("Received bus message")
			switch msg.Type {
			case gst.ERROR:
				debug, err := msg.ParseAsError()
				panicIfError(err)
				logger.Errorw(debug, "func", "StartMsgBus")
				return
			case gst.END_OF_STREAM:
				logger.Debugw("End of camera stream")
				return

			default:
				logger.DPanicw("Unknown message type received", "msgType", msg.Type)

			}

		}
	}
}

package main

//type CameraStream struct {
//	Name     string
//	Host     string
//	Port     int
//	Path     string
//	Username string
//	Password string
//
//	Pipeline gst.Pipeline
//
//	targetTrackSinks map[string]gst.Bin
//}
//
//func NewCameraStream(Id string, host string, port int, path string, username string, password string, logger *zap.SugaredLogger) CameraStream {
//
//	//if ok := srcPipeline.AddElement(source); !ok {
//	//	panic("could not add source element")
//	//}
//
//	//if ok := srcPipeline.AddElement(queue); !ok {
//	//	panic("could not add queue element")
//	//}
//
//	//if ok := srcPipeline.AddElement(depay); !ok {
//	//	panic("could not add depayloader element")
//	//}
//
//	//if ok := srcPipeline.AddElement(parse); !ok {
//	//	panic("could not add parser element")
//	//}
//
//	if ok := srcPipeline.AddElement(test); !ok {
//		panic("could not add test element")
//	}
//
//	if ok := srcPipeline.AddElement(scale); !ok {
//		panic("could not add scale element")
//	}
//
//	if ok := srcPipeline.AddElement(filter); !ok {
//		panic("could not add filter element")
//	}
//
//	//if ok := srcPipeline.AddElement(decode); !ok {
//	//	panic("could not add H265 decoder element")
//	//}
//
//	if ok := srcPipeline.AddElement(vp8encode); !ok {
//		panic("could not add VP8 encoder element")
//	}
//
//	if ok := srcPipeline.AddElement(tee); !ok {
//		panic("could not add tee element")
//	}
//
//	//err = gst.LinkElements(queue, depay)
//	//panicIfError(err)
//	//err = gst.LinkElements(depay, parse)
//	//panicIfError(err)
//	//err = gst.LinkElements(parse, decode)
//	//panicIfError(err)
//	//err = gst.LinkElements(decode, scale)
//	err = gst.LinkElements(test, scale)
//	panicIfError(err)
//	err = gst.LinkElements(scale, filter)
//	panicIfError(err)
//	err = gst.LinkElements(filter, vp8encode)
//	panicIfError(err)
//	err = gst.LinkElements(vp8encode, tee)
//	panicIfError(err)
//
//	//source.OnPadAdded(func(newPad gst.Pad) {
//	//	logger := logger.Named("OnPadAdded").With("element", source.Name())
//	//	logger.Debugw("Received new pad")
//	//	format, err := newPad.Caps().Format(0)
//	//	panicIfError(err)
//	//
//	//	encoding, err := format.QueryStringProperty("encoding-Id")
//	//	panicIfError(err)
//	//
//	//	if encoding != "H265" {
//	//		logger.Debugw("Ignoring unknown codec", "codec", encoding)
//	//		return
//	//	}
//	//
//	//	sinkPad, err := queue.QueryPadByName("sink")
//	//	panicIfError(err)
//	//
//	//	logger.Debugw("Linking new source pad")
//	//	err = gst.LinkPads(newPad, sinkPad)
//	//	panicIfError(err)
//	//
//	//})
//
//	return CameraStream{
//		Name:             Id,
//		Host:             host,
//		Port:             port,
//		Path:             path,
//		Username:         username,
//		Password:         password,
//		Pipeline:         &srcPipeline,
//		targetTrackSinks: make(map[string]gst.Bin),
//	}
//}
//
//func (c *CameraStream) AddOutputTrack(track *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) {
//
//	logger.Debugw("Adding trackSinkBin to srcPipeline", "srcPipeline", c.Pipeline.Name(), "trackSinkBin", trackSinkBin.Name())
//	if ok := c.Pipeline.AddElement(&trackSinkBin); !ok {
//		logger.Panicw("Could not add track sink bin to srcPipeline", c.Name, "srcPipeline", c.Pipeline.Name(), "trackSinkbin", trackSinkBin.Name())
//	}
//
//	logger.Debugw("Getting tee element from srcPipeline", "srcPipeline", c.Pipeline.Name())
//	tee, ok := c.Pipeline.GetElement(fmt.Sprintf("%s-tee", c.Name))
//	if !ok {
//		logger.Panicw("Could not get tee element for srcPipeline", "srcPipeline", c.Pipeline.Name())
//	}
//
//	logger.Debugw("Linking elements to in trackSinkbin", "trackSinkBin", trackSinkBin.Name())
//
//	logger.Debugw("Linking srcPipeline tee to trackSinkBin", "srcPipeline", c.Pipeline.Name(), "tee", tee.Name(), "trackSinkBin", trackSinkBin.Name())
//	err = gst.LinkElements(tee, &trackSinkBin)
//	panicIfError(err)
//
//	if c.Pipeline.State() != gst.PLAYING {
//		logger.Debugw("Starting srcPipeline", "srcPipeline", c.Pipeline.Name())
//		err := c.Pipeline.SetState(gst.PLAYING)
//		panicIfError(err)
//	} else {
//		logger.Debugw("Starting elements in already playing srcPipeline", "srcPipeline", c.Pipeline.Name())
//		err := trackSinkBin.SetState(gst.PLAYING)
//		panicIfError(err)
//
//	}
//
//	c.targetTrackSinks[track.ID()] = &trackSinkBin
//}
//
//func (c *CameraStream) RemoveOutputTrack(track *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) error {
//	logger = logger.Named("RemoveOutputTrack").With("track", track.ID(), "camera", c.Name)
//	trackSinkBin, ok := c.targetTrackSinks[track.ID()]
//
//	if !ok {
//		return fmt.Errorf("Track with ID '%s' is not registered with stream '%s'", track.ID(), c.Name)
//	}
//
//	if len(c.targetTrackSinks) == 1 {
//		logger.Debugw("Stopping srcPipeline", "srcPipeline", c.Pipeline.Name())
//		err := c.Pipeline.SetState(gst.READY)
//		panicIfError(err)
//	}
//
//	logger.Debugw("Unloading track sink bin", "bin", trackSinkBin.Name())
//	err := trackSinkBin.SetState(gst.NULL)
//	panicIfError(err)
//
//	logger.Debugw("Removing trackSinkBin from camera stream srcPipeline", "srcPipeline", c.Pipeline.Name(), "sink", trackSinkBin.Name(), "queue", trackSinkBin.Name())
//	c.Pipeline.RemoveElement(trackSinkBin)
//
//	delete(c.targetTrackSinks, track.ID())
//
//	return nil
//}

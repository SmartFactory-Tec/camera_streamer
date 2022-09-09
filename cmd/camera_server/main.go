package main

import (
	"camera_server/pkg/gst"
	"camera_server/pkg/gst/elements"
	"fmt"
	"time"
)

// TODO the gst library is most likely full of memory leaks, fix

func main() {
	gst.Init()

	//source, err := elements.NewRtspSource("testSource", "rtsp://wowzaec2demo.streamlock.net/vod/mp4:BigBuckBunny_115k.mp4")
	source, err := elements.NewRtspSource("testSource", "rtsp://admin:L2793C70@10.22.240.53:554/cam/realmonitor?channel=1&subtype=0&proto=Onvif")
	panicIfError(err)

	queue, err := elements.NewQueue("testQueue")
	panicIfError(err)

	depay, err := elements.NewRtpH265Depay("testDepay")
	panicIfError(err)

	parse, err := elements.NewH265Parse("testParse")
	panicIfError(err)

	decode, err := elements.NewAvDecH265("testDec")
	panicIfError(err)

	sink, err := elements.NewAutoVideoSink("testSink")
	panicIfError(err)

	pipeline, err := gst.NewGstPipeline("test-pipeline")
	panicIfError(err)

	pipeline.AddElement(source)
	pipeline.AddElement(queue)
	pipeline.AddElement(depay)
	pipeline.AddElement(parse)
	pipeline.AddElement(decode)
	pipeline.AddElement(sink)

	err = gst.LinkElements(queue, depay)
	panicIfError(err)
	err = gst.LinkElements(depay, parse)
	panicIfError(err)
	err = gst.LinkElements(parse, decode)
	panicIfError(err)
	err = gst.LinkElements(decode, sink)
	panicIfError(err)

	padAddedHandler := func(newPad gst.Pad) {
		format, err := (newPad).Format(0)
		panicIfError(err)

		encoding, err := format.QueryStringProperty("encoding-name")
		panicIfError(err)

		if encoding != "H265" {
			return
		}

		sinkPad, err := queue.QueryPadByName("sink")
		panicIfError(err)

		err = gst.LinkPads(newPad, &sinkPad)
		panicIfError(err)

		println("Linked pads!")

	}

	source.OnPadAdded(padAddedHandler)

	time.Sleep(500 * time.Millisecond)

	err = pipeline.SetState(gst.PLAYING)
	panicIfError(err)

	bus, err := pipeline.Bus()
	panicIfError(err)

	for {
		msg, err := bus.PopMessageWithFilter(gst.ERROR | gst.END_OF_STREAM)
		// If there's an error, there's no message to process
		if err == nil {

			switch msg.Type {
			case gst.ERROR:
				println("Error, exiting...")
				debug, err := msg.ParseAsError()
				panicIfError(err)
				println(debug)
				return
			case gst.END_OF_STREAM:
				println("end of stream")
				return

			default:
				panic(fmt.Errorf("unknown message type %i", msg.Type))

			}

		}
		time.Sleep(25 * time.Millisecond)
	}

}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

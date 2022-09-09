package elements

import (
	"camera_server/pkg/gst"
)

type AutoVideoSink struct {
	*gst.BaseElement
}

func NewAutoVideoSink(name string) (AutoVideoSink, error) {
	createdElement, err := gst.NewGstElement("autovideosink", name)

	if err != nil {
		return AutoVideoSink{}, err
	}

	return AutoVideoSink{&createdElement}, nil
}

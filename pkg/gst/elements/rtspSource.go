package elements

import (
	"camera_server/pkg/gst"
)

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type RtspSource struct {
	location string
	*gst.BaseElement
}

func NewRtspSource(name string, location string) (RtspSource, error) {
	createdElement, err := gst.NewGstElement("rtspsrc", name)

	if err != nil {
		return RtspSource{}, err
	}

	createdElement.SetStringProperty("location", location)

	return RtspSource{location: location, BaseElement: &createdElement}, nil
}

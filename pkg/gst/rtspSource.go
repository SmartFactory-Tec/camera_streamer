package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type RtspSource struct {
	Element
}

func NewRtspSource(name string, location string) (*RtspSource, error) {
	element, err := makeElement(name, "rtspsrc")

	if err != nil {
		return nil, err
	}

	rtspSource := RtspSource{element}
	enableGarbageCollection(&rtspSource)

	rtspSource.SetProperty("location", location)

	return &rtspSource, nil
}

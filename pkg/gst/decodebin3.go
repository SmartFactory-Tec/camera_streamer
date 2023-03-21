package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include <gst/gst.h>

extern void newSampleHandler(GstElement *, long);
*/
import "C"

type DecodeBin3 struct {
	Element
}

func NewDecodeBin3(name string) (*DecodeBin3, error) {
	element, err := makeElement(name, "decodebin3")

	if err != nil {
		return nil, err
	}

	decodeBin3 := DecodeBin3{element}
	enableGarbageCollection(&decodeBin3)

	return &decodeBin3, nil
}

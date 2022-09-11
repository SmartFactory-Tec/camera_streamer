package gst

/*
#cgo pkg-config: gstreamer-1.0
#include <gst/gst.h>
*/
import "C"

type Sample interface {
	Buffer() Buffer
	baseSample() *BaseSample
}

type BaseSample struct {
	gstSample *C.GstSample
}

func (b *BaseSample) baseSample() *BaseSample {
	return b
}

func (b *BaseSample) Buffer() Buffer {
	return &BaseBuffer{
		C.gst_sample_get_buffer(b.gstSample),
	}
}

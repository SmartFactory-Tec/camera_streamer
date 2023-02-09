package gst

/*
#cgo pkg-config: gstreamer-1.0
#include <gst/gst.h>
*/
import "C"

type Sample struct {
	gstSample *C.GstSample
}

func wrapSample(gstSample *C.GstSample) Sample {
	return Sample{
		gstSample,
	}
}

func (s *Sample) Buffer() *Buffer {
	// not transferring ownership, not enabling gc
	gstBuffer := C.gst_sample_get_buffer(s.gstSample)
	buffer := wrapGstBuffer(gstBuffer)
	return &buffer
}

func (s *Sample) ref() {
	C.gst_sample_ref(s.gstSample)
}

func (s *Sample) unref() {
	C.gst_sample_unref(s.gstSample)
}

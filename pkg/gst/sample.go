package gst

/*
#cgo pkg-config: gstreamer-1.0
#include <gst/gst.h>
*/
import "C"
import "runtime"

type Sample struct {
	gstSample *C.GstSample
}

func newSample(gstSample *C.GstSample) Sample {
	sample := Sample{gstSample}
	// Unref sample when GC runs
	runtime.SetFinalizer(&sample, func(sample *Sample) {
		C.gst_sample_unref(sample.gstSample)
	})
	return sample
}

func (b *Sample) Buffer() Buffer {
	return Buffer{
		C.gst_sample_get_buffer(b.gstSample),
	}
}

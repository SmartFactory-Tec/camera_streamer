package gst

/*
#cgo pkg-config: gstreamer-1.0
#include <gst/gst.h>
*/
import "C"
import (
	"time"
	"unsafe"
)

type Buffer struct {
	gstBuffer *C.GstBuffer
}

func (b *Buffer) Bytes() []byte {
	var (
		bufferCopy C.gpointer
		copySize   C.gsize
	)
	C.gst_buffer_extract_dup(b.gstBuffer, 0, C.gst_buffer_get_size(b.gstBuffer), &bufferCopy, &copySize)
	return C.GoBytes(unsafe.Pointer(bufferCopy), C.int(copySize))
}

func (b *Buffer) Duration() time.Duration {
	return time.Duration(b.gstBuffer.duration)
}

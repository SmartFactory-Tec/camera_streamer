package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import "unsafe"

type PadTemplate struct {
	gstPadTemplate *C.GstPadTemplate
	Object
}

func wrapGstPadTemplate(gstPadTemplate *C.GstPadTemplate) PadTemplate {
	return PadTemplate{
		gstPadTemplate,
		wrapGstObject((*C.GstObject)(unsafe.Pointer(gstPadTemplate))),
	}
}

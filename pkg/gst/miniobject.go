package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type MiniObject struct {
	gstMiniObject *C.GstMiniObject
}

func wrapGstMiniObject(gstMiniObject *C.GstMiniObject) MiniObject {
	return MiniObject{
		gstMiniObject,
	}
}

func (m *MiniObject) ref() {
	C.gst_mini_object_ref(m.gstMiniObject)
}

func (m *MiniObject) unref() {
	C.gst_mini_object_unref(m.gstMiniObject)
}

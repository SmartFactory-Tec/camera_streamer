package gst

/*
   #cgo pkg-config: gstreamer-1.0

   #include <gst/gst.h>
   #include "object.h"
*/
import "C"
import (
	"unsafe"
)

type Object struct {
	gstObject *C.GstObject
}

func wrapGstObject(pointer *C.GstObject) Object {
	return Object{pointer}
}

func (g *Object) ref() {
	C.gst_object_ref((C.gpointer)(unsafe.Pointer(g.gstObject)))
}

func (g *Object) unref() {
	C.gst_object_unref((C.gpointer)(unsafe.Pointer(g.gstObject)))
}

func (g *Object) Name() string {
	cStr := C.gst_object_get_name(g.gstObject)
	defer C.g_free((C.gpointer)(unsafe.Pointer(cStr)))

	return C.GoString(cStr)
}

func (g *Object) SetProperty(name string, value any) {
	switch value := value.(type) {
	case string:
		C.gst_set_string_property(g.gstObject, C.CString(name), C.CString(value))
		break
	case bool:
		cValue := C.bool(value)
		C.gst_set_bool_property(g.gstObject, C.CString(name), &cValue)
		break
	case int:
		C.gst_set_int_property(g.gstObject, C.CString(name), C.int(value))
		break
	case *Caps:
		C.gst_set_caps_property(g.gstObject, C.CString(name), value.gstCaps)
	default:
		panic("Unsupported type for element property!")
	}
}

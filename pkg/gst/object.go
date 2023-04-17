package gst

/*
   #cgo pkg-config: gstreamer-1.0

   #include <gst/gst.h>

GType g_type_string() {
	return G_TYPE_STRING;
}

GType g_type_boolean() {
	return G_TYPE_BOOLEAN;
}

GType g_type_float() {
	return G_TYPE_FLOAT;
}

GType g_type_int() {
	return G_TYPE_INT;
}

GType g_type_caps() {
	return GST_TYPE_CAPS;
}
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
	gValue := C.GValue{}
	switch value := value.(type) {
	case string:
		C.g_value_init(&gValue, C.g_type_string())
		C.g_value_set_string(&gValue, C.CString(value))
		break
	case bool:
		C.g_value_init(&gValue, C.g_type_boolean())
		var gBool C.gboolean
		if value {
			gBool = 1
		} else {
			gBool = 0
		}
		C.g_value_set_boolean(&gValue, gBool)
		break
	case int:
		C.g_value_init(&gValue, C.g_type_int())
		cInt := C.int(value)
		C.g_value_set_int(&gValue, cInt)
		break
	case float32:
		C.g_value_init(&gValue, C.g_type_float())
		cFloat := C.float(value)
		C.g_value_set_float(&gValue, cFloat)
	case *Caps:
		C.g_value_init(&gValue, C.g_type_caps())
		C.gst_value_set_caps(&gValue, value.gstCaps)
	default:
		panic("Unsupported type for element property!")
	}

	C.g_object_set_property(&g.gstObject.object, C.CString(name), &gValue)
}

package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
)

type Format struct {
	gstStructure *C.GstStructure
	Name         string
}

func (f *Format) QueryStringProperty(propertyName string) (string, error) {
	value := C.gst_structure_get_string(f.gstStructure, C.CString(propertyName))

	if value == nil {
		return "", fmt.Errorf("property '%s' not found", propertyName)
	}

	return C.GoString(value), nil
}

// TODO implement accesor for format properties

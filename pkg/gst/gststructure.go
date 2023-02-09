package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
)

type Structure struct {
	gstStructure *C.GstStructure
}

func wrapGstStructure(gstStructure *C.GstStructure) Structure {
	return Structure{
		gstStructure,
	}
}

func (s *Structure) Name() string {
	// In theory this doesn't require a free
	return C.GoString(C.gst_structure_get_name(s.gstStructure))
}

func (s *Structure) QueryStringProperty(propertyName string) (string, error) {
	value := C.gst_structure_get_string(s.gstStructure, C.CString(propertyName))

	if value == nil {
		return "", fmt.Errorf("property '%s' not found", propertyName)
	}

	return C.GoString(value), nil
}

// TODO implement getter for format properties

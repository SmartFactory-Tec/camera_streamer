package gst

/*
   #cgo pkg-config: gstreamer-1.0

   #include <gst/gst.h>
*/
import "C"
import (
	"fmt"
)

type Caps struct {
	gstCaps *C.GstCaps
	MiniObject
}

func NewCapsFromString(descriptor string) (*Caps, error) {
	gstCaps := C.gst_caps_from_string(C.CString(descriptor))

	if gstCaps == nil {
		return nil, fmt.Errorf("error creating caps from string: %s", descriptor)
	}

	caps := wrapGstCaps(gstCaps)
	enableGarbageCollection(&caps)

	return &caps, nil

}

func wrapGstCaps(gstCaps *C.GstCaps) Caps {
	return Caps{
		gstCaps,
		wrapGstMiniObject(&gstCaps.mini_object),
	}
}

func (c *Caps) Format(index int) (*Structure, error) {
	gstStructure := C.gst_caps_get_structure(c.gstCaps, C.uint(index))

	if gstStructure == nil {
		return nil, fmt.Errorf("error accessing format structure for caps at index '%d'", index)
	}

	structure := wrapGstStructure(gstStructure)

	return &structure, nil
}

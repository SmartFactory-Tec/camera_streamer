package gst

/*
   #cgo pkg-config: gstreamer-1.0

   #include <gst/gst.h>
*/
import "C"
import (
	"fmt"
)

type Caps interface {
	Format(index int) (Format, error)
	capsBase() *BaseCaps
}

type BaseCaps struct {
	gstCaps *C.GstCaps
}

func NewBaseCapsFromString(descriptor string) (BaseCaps, error) {
	gstCaps := C.gst_caps_from_string(C.CString(descriptor))

	if gstCaps == nil {
		return BaseCaps{}, fmt.Errorf("error creating caps from string: %s", descriptor)
	}

	return BaseCaps{gstCaps}, nil

}

func (b *BaseCaps) Format(index int) (Format, error) {
	//if caps == nil {
	//	return Format{}, fmt.Errorf("Error accesing caps")
	//}

	format := Format{
		gstStructure: C.gst_caps_get_structure(b.gstCaps, C.uint(index)),
	}

	if format.gstStructure == nil {
		return Format{}, fmt.Errorf("Error accessing format structure for caps at index '%i'.", index)
	}

	format.Name = C.GoString(C.gst_structure_get_name(format.gstStructure))

	return format, nil
}

func (b *BaseCaps) capsBase() *BaseCaps {
	return b
}

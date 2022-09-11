package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
)

type Pad interface {
	Format(index int) (Format, error)
	padBase() *BasePad
}

type BasePad struct {
	gstPad *C.GstPad
}

func (p *BasePad) padBase() *BasePad {
	return p
}

func (p *BasePad) Format(index int) (Format, error) {
	caps := C.gst_pad_get_current_caps(p.gstPad)

	if caps == nil {
		return Format{}, fmt.Errorf("Error accesing caps")
	}

	format := Format{
		gstStructure: C.gst_caps_get_structure(caps, C.uint(index)),
	}

	if format.gstStructure == nil {
		return Format{}, fmt.Errorf("Error accessing format structure for caps at padAddedIndex '%i'.", index)
	}

	format.Name = C.GoString(C.gst_structure_get_name(format.gstStructure))

	return format, nil
}

func LinkPads(first Pad, second Pad) error {
	ret := C.gst_pad_link(first.padBase().gstPad, second.padBase().gstPad)

	switch ret {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("pads have no common grandparent")
	case -2:
		return fmt.Errorf("pads were already linked")
	case -3:
		return fmt.Errorf("pads have wrong direction")
	case -4:
		return fmt.Errorf("pads do not have common format")
	case -5:
		return fmt.Errorf("pads cannot cooperate in scheduling")
	case -6:
		return fmt.Errorf("pads refused to be linked")
	default:
		return fmt.Errorf("unknown error")
	}

}

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

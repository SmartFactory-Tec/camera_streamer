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
	Caps() Caps
	padBase() *BasePad
}

type BasePad struct {
	gstPad *C.GstPad
}

func (p *BasePad) Caps() Caps {
	gstCaps := C.gst_pad_get_current_caps(p.gstPad)

	return &BaseCaps{
		gstCaps,
	}

}

func (p *BasePad) padBase() *BasePad {
	return p
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

package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Pad struct {
	gstPad *C.GstPad
	Object
}

func wrapPad(gstPad *C.GstPad) Pad {
	return Pad{
		gstPad,
		wrapGstObject((*C.GstObject)(unsafe.Pointer(gstPad))),
	}
}

func (p *Pad) Caps() (*Caps, error) {
	gstCaps := C.gst_pad_get_current_caps(p.gstPad)

	if gstCaps == nil {
		return nil, fmt.Errorf("could not get caps for pad")
	}

	caps := wrapGstCaps(gstCaps)
	enableGarbageCollection(&caps)

	return &caps, nil

}

func (p *Pad) PadTemplate() (*PadTemplate, error) {
	gstPadTemplate := C.gst_pad_get_pad_template(p.gstPad)
	if gstPadTemplate == nil {
		return nil, fmt.Errorf("coulld not get pad template for pad")
	}

	padTemplate := wrapGstPadTemplate(gstPadTemplate)
	enableGarbageCollection(&padTemplate)

	return &padTemplate, nil
}

func LinkPads(first *Pad, second *Pad) error {
	ret := C.gst_pad_link(first.gstPad, second.gstPad)

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

func UnlinkPads(first *Pad, second *Pad) error {
	ok := int(C.gst_pad_unlink(first.gstPad, second.gstPad)) != 0

	if !ok {
		return fmt.Errorf("could not unlink pads")
	}
	return nil

}

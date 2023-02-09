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

type GhostPad struct {
	gstGhostPad *C.GstGhostPad
	Pad
}

func NewGhostPad(name string, original *Pad) (*GhostPad, error) {
	// TODO original must be unlinked
	gstPad := C.gst_ghost_pad_new(C.CString(name), original.gstPad)
	if gstPad == nil {
		return nil, fmt.Errorf("could not create ghost pad")
	}

	ghostPad := wrapGhostPad((*C.GstGhostPad)(unsafe.Pointer(gstPad)))
	enableGarbageCollection(&ghostPad)

	return &ghostPad, nil
}

func wrapGhostPad(gstGhostPad *C.GstGhostPad) GhostPad {
	return GhostPad{
		gstGhostPad,
		wrapPad((*C.GstPad)(unsafe.Pointer(gstGhostPad))),
	}
}

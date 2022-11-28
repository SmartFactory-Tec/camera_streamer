package gst

/*
   #cgo pkg-config: gstreamer-1.0

   #include <gst/gst.h>
*/
import "C"
import "fmt"

type GhostPad struct {
	referencedPad Pad
	*BasePad
}

func NewGhostPad(name string, original Pad) (GhostPad, error) {
	// TODO original must be unlinked
	gstGhostPad := C.gst_ghost_pad_new(C.CString(name), original.padBase().gstPad)
	if gstGhostPad == nil {
		return GhostPad{}, fmt.Errorf("could not create ghost pad")
	}
	return GhostPad{original, &BasePad{gstGhostPad}}, nil
}

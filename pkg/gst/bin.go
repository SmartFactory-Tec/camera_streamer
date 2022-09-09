package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type Bin interface {
	AddElement(Element)
	binBase() *BaseBin
	Element
}

type BaseBin struct {
	BaseElement
	gstBin *C.GstBin
}

func (b *BaseBin) binBase() *BaseBin {
	return b
}

func (b *BaseBin) AddElement(element Element) {
	C.gst_bin_add(b.gstBin, element.elementBase().gstElement)
}

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

type Bin struct {
	gstBin *C.GstBin
	Element
}

func NewBin(name string) (*Bin, error) {
	gstBin := C.gst_bin_new(C.CString(name))

	if gstBin == nil {
		return nil, fmt.Errorf("error while creating bin with name '%s'", name)
	}

	bin := wrapBin((*C.GstBin)(unsafe.Pointer(gstBin)))
	enableGarbageCollection(&bin)

	return &bin, nil
}

func wrapBin(gstBin *C.GstBin) Bin {
	return Bin{
		gstBin,
		wrapGstElement((*C.GstElement)(unsafe.Pointer(gstBin))),
	}
}

func (b *Bin) AddElement(element elementCastable) bool {
	return C.gst_bin_add(b.gstBin, element.element().gstElement) != 0
}

func (b *Bin) RemoveElement(element elementCastable) bool {
	return C.gst_bin_remove(b.gstBin, element.element().gstElement) != 0
}

func (b *Bin) GetElement(name string) (*Element, bool) {
	gstElement := C.gst_bin_get_by_name(b.gstBin, C.CString(name))

	if gstElement == nil {
		return nil, false
	}

	// This is a new reference for the element, it should be safe to wrap it and add the finalizer
	element := wrapGstElement(gstElement)
	enableGarbageCollection(&element)

	return &element, true
}

//func (b *Bin) AddSrcElement(element Element) bool {
//	if ok := b.AddElement(element); !ok {
//		return false
//	}
//
//	pad, err := element.GetPad("src")
//	if err != nil {
//		return false
//	}
//
//	ghostSrc, err := NewGhostPad("src", pad)
//	if err != nil {
//		return false
//	}
//
//	if err := b.AddPad(ghostSrc); err != nil {
//		return false
//	}
//
//	return true
//}
//
//func (b *BaseBin) AddSinkElement(element Element) bool {
//	if ok := b.AddElement(element); !ok {
//		return false
//	}
//
//	pad, err := element.GetPad("sink")
//	if err != nil {
//		return false
//	}
//
//	ghostSrc, err := NewGhostPad("sink", pad)
//	if err != nil {
//		return false
//	}
//
//	if err := b.AddPad(ghostSrc); err != nil {
//		return false
//	}
//
//	return true
//}

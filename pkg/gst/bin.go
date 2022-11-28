package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

type Bin interface {
	AddElement(Element) bool
	RemoveElement(Element) bool
	GetElement(string) (Element, bool)
	binBase() *BaseBin
	Element
}

type BaseBin struct {
	BaseElement
	Elements *sync.Map
	gstBin   *C.GstBin
}

func NewBaseBin(name string) (BaseBin, error) {
	gstBin := C.gst_bin_new(C.CString(name))

	if gstBin == nil {
		return BaseBin{}, fmt.Errorf("error while creating bin with name '%s'", name)
	}

	createdBin := BaseBin{
		gstBin: (*C.GstBin)(unsafe.Pointer(gstBin)),
		BaseElement: BaseElement{
			gstElement:   (*C.GstElement)(unsafe.Pointer(gstBin)),
			elementState: NULL,
			elementType:  "bin",
		},
		Elements: new(sync.Map),
	}

	return createdBin, nil
}

func (b *BaseBin) binBase() *BaseBin {
	return b
}

func (b *BaseBin) AddElement(element Element) bool {
	b.Elements.Store(element.Name(), element)
	return C.gst_bin_add(b.gstBin, element.elementBase().gstElement) != 0
}

func (b *BaseBin) RemoveElement(element Element) bool {
	b.Elements.Delete(element.Name())
	return C.gst_bin_remove(b.gstBin, element.elementBase().gstElement) != 0
}

func (b *BaseBin) GetElement(name string) (Element, bool) {
	if value, ok := b.Elements.Load(name); !ok {
		return nil, false
	} else {
		return value.(Element), true
	}
}

func (b *BaseBin) AddSrcElement(element Element) bool {
	if ok := b.AddElement(element); !ok {
		return false
	}

	pad, err := element.QueryPadByName("src")
	if err != nil {
		return false
	}

	ghostSrc, err := NewGhostPad("src", pad)
	if err != nil {
		return false
	}

	if err := b.AddPad(ghostSrc); err != nil {
		return false
	}

	return true
}

func (b *BaseBin) AddSinkElement(element Element) bool {
	if ok := b.AddElement(element); !ok {
		return false
	}

	pad, err := element.QueryPadByName("sink")
	if err != nil {
		return false
	}

	ghostSrc, err := NewGhostPad("sink", pad)
	if err != nil {
		return false
	}

	if err := b.AddPad(ghostSrc); err != nil {
		return false
	}

	return true
}

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

type Bin interface {
	AddElement(Element) bool
	RemoveElement(Element) bool
	GetElement(string) (Element, bool)
	GetElementCount() int
	binBase() *BaseBin
	Element
}

type BaseBin struct {
	BaseElement
	Elements map[string]Element
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
		Elements: make(map[string]Element),
	}

	return createdBin, nil
}

func (b *BaseBin) binBase() *BaseBin {
	return b
}

func (b *BaseBin) AddElement(element Element) bool {
	b.Elements[element.Name()] = element
	return C.gst_bin_add(b.gstBin, element.elementBase().gstElement) != 0
}

func (b *BaseBin) RemoveElement(element Element) bool {
	return C.gst_bin_remove(b.gstBin, element.elementBase().gstElement) != 0
}

func (b *BaseBin) GetElement(name string) (Element, bool) {
	if element, ok := b.Elements[name]; ok {
		return element, true
	} else {
		return nil, false
	}
}

func (b *BaseBin) GetElementCount() int {
	return len(b.Elements)
}

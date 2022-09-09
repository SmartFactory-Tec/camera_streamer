package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
#include "element.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type Element interface {
	Name() string
	Type() string
	SetState(state ElementState) error
	elementBase() *BaseElement
}

type BaseElement struct {
	elementType string
	gstElement  *C.GstElement
}

func (g *BaseElement) Name() string {
	name := C.gst_object_get_name((*C.GstObject)(unsafe.Pointer(g.gstElement)))
	return C.GoString(name)
}

func (g *BaseElement) Type() string {
	// TODO find a way to dynamically get the type instead of storing a copy
	return g.elementType
}

func (g *BaseElement) elementBase() *BaseElement {
	return g
}

func NewGstElement(elementType string, name string) (BaseElement, error) {
	gstElementFactory := C.gst_element_factory_find(C.CString(elementType))

	if gstElementFactory == nil {
		return BaseElement{}, fmt.Errorf("error creating element of type '%s' with Name '%s', no such type found", elementType, name)
	}

	newGstElement := C.gst_element_factory_make(C.CString(elementType), C.CString(name))

	if newGstElement == nil {
		return BaseElement{}, fmt.Errorf("error creating element of type '%s', with Name '%s'", elementType, name)
	}

	return BaseElement{elementType, newGstElement}, nil
}

func (g *BaseElement) QueryPadByName(name string) (BasePad, error) {
	foundPad := C.gst_element_get_static_pad(g.gstElement, C.CString(name))
	if foundPad == nil {
		return BasePad{}, fmt.Errorf("Error finding pad with Name '%s', on element '%s'['%s']", name, g.Name(), g.Type())
	}
	return BasePad{foundPad}, nil
}

func (g *BaseElement) SetStringProperty(name string, value string) {
	C.gst_set_string_property(g.gstElement, C.CString(name), C.CString(value))
}

// LinkElements wrapper for C function to link two elements in a pipeline
func LinkElements(first Element, second Element) error {
	result := C.gst_element_link(first.elementBase().gstElement, second.elementBase().gstElement)

	if result == 0 {
		return fmt.Errorf("failed to link elements %s[%s] and %s[%s]", first.Name(),
			first.Type(), second.Name(), second.Type())
	}

	return nil
}

type ElementState int

const (
	PLAYING ElementState = iota
)

func (g *BaseElement) SetState(state ElementState) error {
	switch state {
	case PLAYING:
		if !C.setStatePlaying(g.gstElement) {
			return fmt.Errorf("could not change to requested state")
		}
	default:
		return fmt.Errorf("unknown state change requested")
	}
	return nil
}

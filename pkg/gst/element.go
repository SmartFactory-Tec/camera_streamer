package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
#include "element.h"
#include "callbacks.h"

extern void padAddedHandler(GstElement*, GstPad*, long);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Interface to enable all structs that embed an element to be used as one
type elementCastable interface {
	element() *Element
}

type Element struct {
	gstElement *C.GstElement
	Object
}

// makeElement creates one of element's subclasses using their respective factories, returning an Element wrapper
// that does not have GC enabled.
func makeElement(name string, elementType string) (Element, error) {
	gstElementFactory := C.gst_element_factory_find(C.CString(elementType))

	if gstElementFactory == nil {
		return Element{}, fmt.Errorf("error creating element of type '%s' with Name '%s', no such type found", elementType, name)
	}

	newGstElement := C.gst_element_factory_make(C.CString(elementType), C.CString(name))

	if newGstElement == nil {
		return Element{}, fmt.Errorf("error creating element of type '%s', with Name '%s'", elementType, name)
	}

	element := wrapGstElement(newGstElement)

	return element, nil
}

// Identity function to enable interface usage
func (e *Element) element() *Element {
	return e
}

func wrapGstElement(gstElement *C.GstElement) Element {
	return Element{
		gstElement,
		wrapGstObject((*C.GstObject)(unsafe.Pointer(gstElement))),
	}
}

func (e *Element) State() ElementState {
	var (
		current C.GstState
		pending C.GstState
	)
	state := (int)(C.gst_element_get_state(e.gstElement, &current, &pending, 0))

	switch state {
	case 1:
		return NULL
	case 2:
		return READY
	case 3:
		return PAUSED
	case 4:
		return PLAYING
	default:
		panic("Element in unknown state!")
	}

}

func (e *Element) Type() string {
	// TODO check if this works
	factory := C.gst_element_get_factory(e.gstElement)

	factoryObject := Object{(*C.GstObject)(unsafe.Pointer(factory))}

	return factoryObject.Name()
}

func (e *Element) GetPad(name string) (*Pad, bool) {
	gstPad := C.gst_element_get_static_pad(e.gstElement, C.CString(name))
	if gstPad == nil {
		return nil, false
	}

	// get_static_path transfers ownership, enable unref on garbage collect
	pad := wrapPad(gstPad)
	enableGarbageCollection(&pad)

	return &pad, true
}

// LinkElements wrapper for C function to link two elements in a pipeline
func LinkElements(first elementCastable, second elementCastable) error {
	result := C.gst_element_link(first.element().gstElement, second.element().gstElement)

	if result == 0 {
		return fmt.Errorf("failed to link elements %s[%s] and %s[%s]", first.element().Name(),
			first.element().Type(), second.element().Name(), second.element().Type())
	}

	return nil
}

type ElementState int

const (
	NULL ElementState = iota
	PLAYING
	READY
	PAUSED
)

func (e *Element) SetState(state ElementState) error {
	switch state {
	case PLAYING:
		if !C.setStatePlaying(e.gstElement) {
			return fmt.Errorf("could not change state to playing")
		}
	case READY:
		if !C.setStateReady(e.gstElement) {
			return fmt.Errorf("could not change state to ready")
		}
	case NULL:
		if !C.setStateNull(e.gstElement) {
			return fmt.Errorf("could not change state to null")
		}
	case PAUSED:
		if !C.setStatePaused(e.gstElement) {
			return fmt.Errorf("could not change state to paused")
		}
	default:
		// This should really not happen
		panic("unknown state requested")
	}
	return nil
}

func (e *Element) AddPad(pad *Pad) error {
	if ret := C.gst_element_add_pad(e.gstElement, pad.gstPad); ret == 0 {
		return fmt.Errorf("could not add pad to element")
	}
	return nil
}

type PadAddedCallback func(pad *Pad)

var (
	padAddedIndex     int64 = 0
	padAddedCallbacks       = make(map[int64]PadAddedCallback)
	padAddedLock      sync.Mutex
)

//export padAddedHandler
func padAddedHandler(_ *C.GstElement, newPad *C.GstPad, callbackID C.long) {
	padAddedLock.Lock()
	defer padAddedLock.Unlock()

	if callback, ok := padAddedCallbacks[int64(callbackID)]; ok {
		pad := wrapPad(newPad)
		enableGarbageCollection(&pad)
		callback(&pad)
	} else {
		panic("callback not found")
	}
}

func (e *Element) OnPadAdded(callback PadAddedCallback) {
	padAddedLock.Lock()
	defer padAddedLock.Unlock()

	padAddedCallbacks[padAddedIndex] = callback
	C.connectSignalHandler(C.CString("pad-added"), e.gstElement, C.padAddedHandler, C.long(padAddedIndex))
	padAddedIndex++
}

func (e *Element) RequestPad(name string) (*Pad, error) {
	gstPad := C.gst_element_request_pad_simple(e.gstElement, C.CString(name))

	if gstPad == nil {
		return nil, fmt.Errorf("could not request pad with name %s", name)
	}

	pad := wrapPad(gstPad)
	enableGarbageCollection(&pad)

	return &pad, nil
}

func (e *Element) ReleaseRequestPad(pad *Pad) {
	C.gst_element_release_request_pad(e.gstElement, pad.gstPad)
}

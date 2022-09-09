package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
#include "callbacks.h"
*/
import "C"
import (
	"sync"
)

var (
	index             int64 = 0
	padAddedCallbacks sync.Map
)

type PadAddedCallback func(pad Pad)

//export padAddedHandler
func padAddedHandler(_ *C.GstElement, newPad *C.GstPad, callbackID C.long) {
	if callback, ok := padAddedCallbacks.Load(int64(callbackID)); ok {
		callback.(PadAddedCallback)(&BasePad{newPad})
	} else {
		panic("callback not found")
	}
}

func (g *BaseElement) OnPadAdded(callback PadAddedCallback) {
	padAddedCallbacks.Store(index, callback)
	C.connectSignalHandler(C.CString("pad-added"), g.gstElement, C.padAddedHandler, C.long(index))
	index++
}

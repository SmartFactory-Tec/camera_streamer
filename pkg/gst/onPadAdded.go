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

type PadAddedCallback func(pad Pad)

var (
	padAddedIndex     int64 = 0
	padAddedCallbacks       = make(map[int64]PadAddedCallback)
	padAddedLock      sync.Mutex
)

//export padAddedHandler
func padAddedHandler(_ *C.GstElement, newPad *C.GstPad, callbackID C.long) {
	padAddedLock.Lock()
	if callback, ok := padAddedCallbacks[int64(callbackID)]; ok {
		callback(&BasePad{newPad})
	} else {
		panic("callback not found")
	}
	padAddedLock.Unlock()
}

func (g *BaseElement) OnPadAdded(callback PadAddedCallback) {
	padAddedLock.Lock()
	padAddedCallbacks[padAddedIndex] = callback
	C.connectSignalHandler(C.CString("pad-added"), g.gstElement, C.padAddedHandler, C.long(padAddedIndex))
	padAddedIndex++
	padAddedLock.Unlock()
}

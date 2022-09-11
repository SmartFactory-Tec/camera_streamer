package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
#include "callbacks.h"
*/
import "C"
import (
	"sync"
	"unsafe"
)

type NewSampleCallback func(newSample Sample)

var (
	newSampleIndex     int64 = 0
	newSampleCallbacks       = make(map[int64]NewSampleCallback)
	newSampleLock      sync.Mutex
)

//export newSampleHandler
func newSampleHandler(element *C.GstElement, callbackID C.long) {
	sample := BaseSample{}

	if C.callSignalByName(element, C.CString("pull-sample"), unsafe.Pointer(&sample.gstSample)); sample.gstSample == nil {
		println("Couldn't pull sample")
		return
	}

	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	if callback, ok := newSampleCallbacks[int64(callbackID)]; ok {
		callback(&sample)
	} else {
		panic("callback not found")
	}
}

func (g *BaseElement) OnNewSample(callback NewSampleCallback) {
	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	newSampleCallbacks[newSampleIndex] = callback
	C.connectSignalHandler(C.CString("new-sample"), g.gstElement, C.newSampleHandler, C.long(newSampleIndex))
	newSampleIndex++
}

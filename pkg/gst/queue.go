package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
#include "callbacks.h"
*/
import "C"
import "sync"

type Queue struct {
	*BaseElement
}

func NewQueue(name string) (Queue, error) {
	createdElement, err := NewGstElement("queue", name)

	if err != nil {
		return Queue{}, err
	}

	return Queue{&createdElement}, nil
}

type OverrunCallback func()

var (
	overrunIndex     int64 = 0
	overrunCallbacks       = make(map[int64]OverrunCallback)
	overrunLock      sync.Mutex
)

//export overrunHandler
func overrunHandler(_ *C.GstElement, callbackID C.long) {
	overrunLock.Lock()
	defer overrunLock.Unlock()

	if callback, ok := overrunCallbacks[int64(callbackID)]; ok {
		callback()
	} else {
		panic("callback not found")
	}
}

func (g *Queue) OnOverrun(callback OverrunCallback) {
	overrunLock.Lock()
	defer overrunLock.Unlock()
	overrunCallbacks[overrunIndex] = callback
	C.connectSignalHandler(C.CString("overrun"), g.gstElement, C.overrunHandler, C.long(overrunIndex))
	overrunIndex++
}

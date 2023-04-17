package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include <gst/gst.h>
#include <gst/app/gstappsink.h>
#include "callbacks.h"

extern void newSampleHandler(GstElement *, long);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

type AppSink struct {
	Element
}

func NewAppSink(name string) (*AppSink, error) {
	element, err := makeElement(name, "appsink")

	if err != nil {
		return nil, err
	}

	appSink := AppSink{element}
	enableGarbageCollection(&appSink)

	return &appSink, nil
}

func (a *AppSink) PullSample() (*Sample, error) {
	var gstSample *C.GstSample = C.gst_app_sink_pull_sample((*C.GstAppSink)(unsafe.Pointer(a.gstElement)))

	if gstSample == nil {
		return nil, fmt.Errorf("appsink is either null or has reached EOS")
	}

	sample := wrapSample(gstSample)
	enableGarbageCollection(&sample)

	return &sample, nil
}

type NewSampleCallback func(newSample *Sample)

var (
	newSampleIndex     int64 = 0
	newSampleCallbacks       = make(map[int64]NewSampleCallback)
	newSampleLock      sync.Mutex
)

//export newSampleHandler
func newSampleHandler(element *C.GstElement, callbackID C.long) {
	var gstSample *C.GstSample

	if C.callSignalByName(element, C.CString("pull-sample"), unsafe.Pointer(&gstSample)); gstSample == nil {
		println("Couldn't pull sample")
		return
	}

	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	if callback, ok := newSampleCallbacks[int64(callbackID)]; ok {
		sample := wrapSample(gstSample)
		enableGarbageCollection(&sample)
		callback(&sample)
	} else {
		panic("callback not found")
	}
}

func (a *AppSink) OnNewSample(callback NewSampleCallback) {
	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	newSampleCallbacks[newSampleIndex] = callback
	C.connectSignalHandler(C.CString("new-sample"), a.gstElement, C.newSampleHandler, C.long(newSampleIndex))
	newSampleIndex++
}

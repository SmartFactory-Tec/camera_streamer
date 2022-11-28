package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include <gst/gst.h>
#include <gst/app/gstappsink.h>
#include "callbacks.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

type AppSink struct {
	*BaseElement
}

func NewAppSink(name string) (AppSink, error) {
	createdElement, err := NewGstElement("appsink", name)

	if err != nil {
		return AppSink{}, err
	}

	return AppSink{&createdElement}, nil
}

func (a *AppSink) PullSample() (Sample, error) {
	var gstSample *C.GstSample = C.gst_app_sink_pull_sample((*C.GstAppSink)(unsafe.Pointer(a.gstElement)))

	if gstSample == nil {
		return Sample{}, fmt.Errorf("appsink is either null or has reached EOS")
	}

	return newSample(gstSample), nil
}

type NewSampleCallback func(newSample Sample)

var (
	newSampleIndex     int64 = 0
	newSampleCallbacks       = make(map[int64]NewSampleCallback)
	newSampleLock      sync.Mutex
)

//export newSampleHandler
func newSampleHandler(element *C.GstElement, callbackID C.long) {
	sample := Sample{}

	if C.callSignalByName(element, C.CString("pull-sample"), unsafe.Pointer(&sample.gstSample)); sample.gstSample == nil {
		println("Couldn't pull sample")
		return
	}

	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	if callback, ok := newSampleCallbacks[int64(callbackID)]; ok {
		callback(sample)
	} else {
		panic("callback not found")
	}
}

func (g *AppSink) OnNewSample(callback NewSampleCallback) {
	newSampleLock.Lock()
	defer newSampleLock.Unlock()

	newSampleCallbacks[newSampleIndex] = callback
	C.connectSignalHandler(C.CString("new-sample"), g.gstElement, C.newSampleHandler, C.long(newSampleIndex))
	newSampleIndex++
}

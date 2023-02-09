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

type Pipeline struct {
	gstPipeline *C.GstPipeline
	Bin
}

func NewGstPipeline(name string) (*Pipeline, error) {
	gstPipeline := C.gst_pipeline_new(C.CString(name))

	if gstPipeline == nil {
		return nil, fmt.Errorf("error creating pipeline with Name '%s'", name)
	}

	pipeline := wrapGstPipeline((*C.GstPipeline)(unsafe.Pointer(gstPipeline)))
	enableGarbageCollection(&pipeline)

	return &pipeline, nil
}

func wrapGstPipeline(gstPipeline *C.GstPipeline) Pipeline {
	return Pipeline{
		gstPipeline,
		wrapBin((*C.GstBin)(unsafe.Pointer(gstPipeline))),
	}
}

func (e *Pipeline) Bus() (*Bus, error) {
	gstBus := C.gst_element_get_bus(e.gstElement)
	if gstBus == nil {
		return nil, fmt.Errorf("error aquiring gstBus for pipeline %s", e.Name())
	}

	bus := wrapGstBus(gstBus)
	enableGarbageCollection(&bus)

	return &bus, nil
}

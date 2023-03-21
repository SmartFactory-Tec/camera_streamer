package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
	"time"
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

func (e *Pipeline) Latency() time.Duration {
	latency := C.gst_pipeline_get_latency(e.gstPipeline)
	return time.Duration(uint64(latency)) * time.Nanosecond
}

func (e *Pipeline) QueryLatency() (time.Duration, time.Duration, error) {
	var min C.GstClockTime
	var max C.GstClockTime
	query := C.gst_query_new_latency()
	result := int(C.gst_element_query(e.gstElement, query))
	if result != 1 {
		return 0, 0, fmt.Errorf("could not perform query")
	}

	C.gst_query_parse_latency(query, nil, &min, &max)

	return time.Duration(min), time.Duration(max), nil
}

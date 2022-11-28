package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

type Pipeline interface {
	Bin
	pipelineBase() *BasePipeline
	Bus() (Bus, error)
}

type BasePipeline struct {
	BaseBin
	gstPipeline *C.GstPipeline
}

func NewGstPipeline(name string) (BasePipeline, error) {
	gstPipeline := C.gst_pipeline_new(C.CString(name))

	if gstPipeline == nil {
		return BasePipeline{}, fmt.Errorf("error creating pipeline with Name '%s'", name)
	}

	createdPipeline := BasePipeline{
		gstPipeline: (*C.GstPipeline)(unsafe.Pointer(gstPipeline)),
		BaseBin: BaseBin{
			gstBin: (*C.GstBin)(unsafe.Pointer(gstPipeline)),
			BaseElement: BaseElement{
				gstElement:   (*C.GstElement)(unsafe.Pointer(gstPipeline)),
				elementState: NULL,
				elementType:  "pipeline",
			},
			Elements: new(sync.Map),
		},
	}
	return createdPipeline, nil
}

func (e *BasePipeline) Bus() (Bus, error) {
	bus := C.gst_element_get_bus(e.gstElement)
	if bus == nil {
		return Bus{}, fmt.Errorf("error aquiring bus for pipeline %s", e.Name())
	}
	return Bus{gstBus: bus}, nil
}

func (e *BasePipeline) pipelineBase() *BasePipeline {
	return e
}

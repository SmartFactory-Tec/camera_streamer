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

type Bus struct {
	gstBus *C.GstBus
	Object
}

func wrapGstBus(gstBus *C.GstBus) Bus {
	return Bus{
		gstBus,
		wrapGstObject((*C.GstObject)(unsafe.Pointer(gstBus))),
	}
}

func (b *Bus) PopMessageWithFilter(filter MessageType) (*Message, error) {
	gstMessage := C.gst_bus_pop_filtered(b.gstBus, C.GstMessageType(filter))

	if gstMessage == nil {
		return nil, fmt.Errorf("no message found")
	}

	message := wrapGstMessage(gstMessage)
	enableGarbageCollection(&message)

	return &message, nil
}

func (b *Bus) PopMessage() (*Message, error) {
	gstMessage := C.gst_bus_pop(b.gstBus)

	if gstMessage == nil {
		return nil, fmt.Errorf("no message found")
	}

	message := wrapGstMessage(gstMessage)
	enableGarbageCollection(&message)

	return &message, nil
}

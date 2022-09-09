package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import "fmt"

type Bus struct {
	gstBus *C.GstBus
}

func (b *Bus) PopMessageWithFilter(filter MessageType) (Message, error) {
	msg := C.gst_bus_pop_filtered(b.gstBus, C.GstMessageType(filter))

	if msg == nil {
		return Message{}, fmt.Errorf("no message found")
	}

	return Message{gstMessage: msg, Type: MessageType(msg._type)}, nil
}

func (b *Bus) PopMessage() (Message, error) {
	msg := C.gst_bus_pop(b.gstBus)

	if msg == nil {
		return Message{}, fmt.Errorf("no message found")
	}

	return Message{gstMessage: msg, Type: MessageType(msg._type)}, nil
}

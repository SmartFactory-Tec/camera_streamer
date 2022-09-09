package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"
import "fmt"

type MessageType int

const (
	END_OF_STREAM = 1 << iota
	ERROR
	WARNING
	INFO
	TAG
	BUFFERING
	STATE_CHANGED
	STATE_DIRTY
	STEP_DONE
	CLOCK_PROVIDE
	CLOCK_LOST
	NEW_CLOCK
	STRUCTURE_CHANGE
	STREAM_STATUS
	APPLICATION
	ELEMENT
	SEGMENT_START
	SEGMENT_DONE
	DURATION_CHANGED
	LATENCY
	ASYNC_START
	ASYNC_DONE
	REQUEST_STATE
	STEP_START
	QUALITY_OF_SERVICE
	PROGRESS
	TABLE_OF_CONTENTS
	RESET_TIME
	STREAM_START
	NEED_CONTEXT
	HAVE_CONTEXT
	EXTENDED
	DEVICE_ADDED         = EXTENDED + 1
	DEVICE_REMOVED       = DEVICE_ADDED + 1
	PROPERTY_NOTIFY      = DEVICE_REMOVED + 1
	STREAM_COLLECTION    = PROPERTY_NOTIFY + 1
	STREAMS_SELECTED     = STREAM_COLLECTION + 1
	REDIRECT             = STREAMS_SELECTED + 1
	DEVICE_CHANGED       = REDIRECT + 1
	INSTANT_RATE_REQUEST = DEVICE_CHANGED + 1
	GST_MESSAGE_ANY      = 1<<32 - 1
)

type Message struct {
	gstMessage *C.GstMessage
	Type       MessageType
}

func (m *Message) ParseAsError() (string, error) {
	if m.Type != ERROR {
		return "", fmt.Errorf("message is not an error")
	}

	var (
		gError      *C.GError
		errorString *C.char
	)

	C.gst_message_parse_error(m.gstMessage, &gError, &errorString)

	return C.GoString(errorString), nil
}

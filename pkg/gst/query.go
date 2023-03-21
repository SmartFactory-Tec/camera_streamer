package gst

/*
#cgo pkg-config: gstreamer-1.0

#include <gst/gst.h>
*/
import "C"

type Query struct {
	query *C.GstQuery
	MiniObject
}

func wrapQuery(gstQuery *C.GstQuery) Query {
	return Query{
		gstQuery,
		wrapGstMiniObject(&gstQuery.mini_object),
	}
}

func NewLatencyQuery() *Query {
	gstQuery := C.gst_query_new_latency()

	query := wrapQuery(gstQuery)
	enableGarbageCollection(&query)

	return &query
}

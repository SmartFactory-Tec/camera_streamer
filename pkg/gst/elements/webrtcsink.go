package elements

import "C"

type WebRtcSampleCallback func(*C.void, C.int, *C.void)

//var callbacks = make(map[string]WebRtcSampleCallback)
//
//type WebRtcSink struct {
//	gst.Element
//}
//
//func NewWebRtcSink(name string, track webrtc.TrackLocalStaticSample) (*WebRtcSink, error) {
//	createdElement, err := gst.NewElement("appsink", name)
//
//	if err != nil {
//		return nil, err
//	}
//
//	createdElement.SetStringProperty("name", name)
//
//	callbacks[name] = (func(*C.void, C.int, *C.void) {
//
//	})
//
//	return &WebRtcSink{Element: *createdElement}, nil
//}
//
////export newSampleHandler
//func newSampleHandler(element *C.BaseElement, *C.void) {
//	C.g_signal_emit_by_name()
//}

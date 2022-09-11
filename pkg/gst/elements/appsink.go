package elements

import "camera_server/pkg/gst"

type AppSink struct {
	*gst.BaseElement
}

func NewAppSink(name string) (AppSink, error) {
	createdElement, err := gst.NewGstElement("appsink", name)

	if err != nil {
		return AppSink{}, err
	}

	return AppSink{&createdElement}, nil
}

package elements

import "camera_server/pkg/gst"

type H264Parse struct {
	*gst.BaseElement
}

func NewH264Parse(name string) (H264Parse, error) {
	createdElement, err := gst.NewGstElement("h264parse", name)

	if err != nil {
		return H264Parse{}, err
	}

	return H264Parse{&createdElement}, nil
}

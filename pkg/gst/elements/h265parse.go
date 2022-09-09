package elements

import "camera_server/pkg/gst"

type H265Parse struct {
	*gst.BaseElement
}

func NewH265Parse(name string) (H265Parse, error) {
	createdElement, err := gst.NewGstElement("h265parse", name)

	if err != nil {
		return H265Parse{}, err
	}

	return H265Parse{&createdElement}, nil
}

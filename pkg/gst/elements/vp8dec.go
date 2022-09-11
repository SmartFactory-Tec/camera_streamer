package elements

import "camera_server/pkg/gst"

type Vp8Dec struct {
	*gst.BaseElement
}

func NewVp8Dec(name string) (Vp8Dec, error) {
	createdElement, err := gst.NewGstElement("vp8dec", name)

	if err != nil {
		return Vp8Dec{}, err
	}

	return Vp8Dec{&createdElement}, nil
}

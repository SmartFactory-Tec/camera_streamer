package elements

import "camera_server/pkg/gst"

type Vp8Enc struct {
	*gst.BaseElement
}

func NewVp8Enc(name string) (Vp8Enc, error) {
	createdElement, err := gst.NewGstElement("vp8enc", name)

	if err != nil {
		return Vp8Enc{}, err
	}

	return Vp8Enc{&createdElement}, nil
}

package elements

import "camera_server/pkg/gst"

type AvDecH265 struct {
	*gst.BaseElement
}

func NewAvDecH265(name string) (AvDecH265, error) {
	createdElement, err := gst.NewGstElement("avdec_h265", name)

	if err != nil {
		return AvDecH265{}, err
	}

	return AvDecH265{&createdElement}, nil
}

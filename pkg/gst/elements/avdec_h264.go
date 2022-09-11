package elements

import "camera_server/pkg/gst"

type AvDecH264 struct {
	*gst.BaseElement
}

func NewAvDecH264(name string) (AvDecH264, error) {
	createdElement, err := gst.NewGstElement("avdec_h264", name)

	if err != nil {
		return AvDecH264{}, err
	}

	return AvDecH264{&createdElement}, nil
}

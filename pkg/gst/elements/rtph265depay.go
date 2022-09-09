package elements

import "camera_server/pkg/gst"

type RtpH265Depay struct {
	*gst.BaseElement
}

func NewRtpH265Depay(name string) (RtpH265Depay, error) {
	createdElement, err := gst.NewGstElement("rtph265depay", name)

	if err != nil {
		return RtpH265Depay{}, err
	}

	return RtpH265Depay{&createdElement}, nil
}

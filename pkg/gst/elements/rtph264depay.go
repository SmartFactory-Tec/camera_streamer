package elements

import "camera_server/pkg/gst"

type RtpH264Depay struct {
	*gst.BaseElement
}

func NewRtpH264Depay(name string) (RtpH264Depay, error) {
	createdElement, err := gst.NewGstElement("rtph264depay", name)

	if err != nil {
		return RtpH264Depay{}, err
	}

	return RtpH264Depay{&createdElement}, nil
}

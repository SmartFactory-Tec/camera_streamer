package gst

type RtpH264Depay struct {
	*BaseElement
}

func NewRtpH264Depay(name string) (RtpH264Depay, error) {
	createdElement, err := NewGstElement("rtph264depay", name)

	if err != nil {
		return RtpH264Depay{}, err
	}

	return RtpH264Depay{&createdElement}, nil
}

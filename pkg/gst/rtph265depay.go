package gst

type RtpH265Depay struct {
	*BaseElement
}

func NewRtpH265Depay(name string) (RtpH265Depay, error) {
	createdElement, err := NewGstElement("rtph265depay", name)

	if err != nil {
		return RtpH265Depay{}, err
	}

	return RtpH265Depay{&createdElement}, nil
}

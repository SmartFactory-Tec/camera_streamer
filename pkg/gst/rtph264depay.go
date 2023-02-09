package gst

type RtpH264Depay struct {
	Element
}

func NewRtpH264Depay(name string) (*RtpH264Depay, error) {
	element, err := makeElement(name, "rtph264depay")

	if err != nil {
		return nil, err
	}

	rtpH264Depay := RtpH264Depay{element}
	enableGarbageCollection(&rtpH264Depay)

	return &rtpH264Depay, nil
}

package gst

type RtpH265Depay struct {
	Element
}

func NewRtpH265Depay(name string) (*RtpH265Depay, error) {
	element, err := makeElement(name, "rtph265depay")

	if err != nil {
		return nil, err
	}

	rtpH265Depay := RtpH265Depay{element}
	enableGarbageCollection(&rtpH265Depay)

	return &rtpH265Depay, nil
}

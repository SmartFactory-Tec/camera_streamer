package gst

type RtpJitterBuffer struct {
	Element
}

func NewRtpJitterBuffer(name string) (*RtpJitterBuffer, error) {
	element, err := makeElement(name, "rtpjitterbuffer")

	if err != nil {
		return nil, err
	}

	rtpJitterBuffer := RtpJitterBuffer{element}
	enableGarbageCollection(&rtpJitterBuffer)

	return &rtpJitterBuffer, nil
}

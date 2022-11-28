package gst

type RtpJitterBuffer struct {
	*BaseElement
}

func NewRtpJitterBuffer(name string) (RtpJitterBuffer, error) {
	createdElement, err := NewGstElement("rtpjitterbuffer", name)

	if err != nil {
		return RtpJitterBuffer{}, err
	}

	return RtpJitterBuffer{&createdElement}, nil
}

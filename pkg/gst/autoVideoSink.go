package gst

import "C"

type AutoVideoSink struct {
	Element
}

func NewAutoVideoSink(name string) (*AutoVideoSink, error) {
	element, err := makeElement(name, "autovideosink")

	if err != nil {
		return nil, err
	}

	autoVideoSink := AutoVideoSink{element}
	enableGarbageCollection(&autoVideoSink)

	return &autoVideoSink, nil
}

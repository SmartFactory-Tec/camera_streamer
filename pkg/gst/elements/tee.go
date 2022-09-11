package elements

import (
	"camera_server/pkg/gst"
)

type Tee struct {
	*gst.BaseElement
}

func NewTee(name string) (Tee, error) {
	createdElement, err := gst.NewGstElement("tee", name)

	if err != nil {
		return Tee{}, err
	}

	return Tee{&createdElement}, nil
}

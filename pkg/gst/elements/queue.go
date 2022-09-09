package elements

import "camera_server/pkg/gst"

type Queue struct {
	*gst.BaseElement
}

func NewQueue(name string) (Queue, error) {
	createdElement, err := gst.NewGstElement("queue", name)

	if err != nil {
		return Queue{}, err
	}

	return Queue{&createdElement}, nil
}

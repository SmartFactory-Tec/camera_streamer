package gst

type Queue2 struct {
	*BaseElement
}

func NewQueue2(name string) (Queue2, error) {
	createdElement, err := NewGstElement("queue2", name)

	if err != nil {
		return Queue2{}, err
	}

	return Queue2{&createdElement}, nil
}

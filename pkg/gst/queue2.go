package gst

type Queue2 struct {
	Element
}

func NewQueue2(name string) (*Queue2, error) {
	element, err := makeElement(name, "queue2")

	if err != nil {
		return nil, err
	}

	queue2 := Queue2{element}
	enableGarbageCollection(&queue2)

	return &queue2, nil
}

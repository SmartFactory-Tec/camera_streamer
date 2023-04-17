package gst

type Multiqueue struct {
	Element
}

func NewMultiqueue(name string) (*Multiqueue, error) {
	element, err := makeElement(name, "multiqueue")

	if err != nil {
		return nil, err
	}

	multiqueue := Multiqueue{element}
	enableGarbageCollection(&multiqueue)

	return &multiqueue, nil
}

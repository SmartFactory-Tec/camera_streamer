package gst

type Tee struct {
	Element
}

func NewTee(name string) (*Tee, error) {
	element, err := makeElement(name, "tee")

	if err != nil {
		return nil, err
	}

	tee := Tee{element}
	enableGarbageCollection(&tee)

	return &tee, nil
}

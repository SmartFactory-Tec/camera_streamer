package gst

type Tee struct {
	*BaseElement
}

func NewTee(name string) (Tee, error) {
	createdElement, err := NewGstElement("tee", name)

	if err != nil {
		return Tee{}, err
	}

	return Tee{&createdElement}, nil
}

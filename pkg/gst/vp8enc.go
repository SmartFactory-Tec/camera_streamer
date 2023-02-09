package gst

type Vp8Enc struct {
	Element
}

func NewVp8Enc(name string) (*Vp8Enc, error) {
	element, err := makeElement(name, "vp8enc")

	if err != nil {
		return nil, err
	}

	vp8Enc := Vp8Enc{element}
	enableGarbageCollection(&vp8Enc)

	return &vp8Enc, nil
}

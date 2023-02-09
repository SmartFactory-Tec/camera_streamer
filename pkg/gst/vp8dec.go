package gst

type Vp8Dec struct {
	Element
}

func NewVp8Dec(name string) (*Vp8Dec, error) {
	element, err := makeElement(name, "vp8dec")

	if err != nil {
		return nil, err
	}

	vp8Dec := Vp8Dec{element}
	enableGarbageCollection(&vp8Dec)

	return &vp8Dec, nil
}

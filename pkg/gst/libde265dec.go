package gst

type LibDe265Dec struct {
	*BaseElement
}

func NewLibDe265Dec(name string) (LibDe265Dec, error) {
	createdElement, err := NewGstElement("libde265dec", name)

	if err != nil {
		return LibDe265Dec{}, err
	}

	return LibDe265Dec{&createdElement}, nil
}

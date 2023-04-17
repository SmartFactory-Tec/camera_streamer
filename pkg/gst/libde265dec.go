package gst

type LibDe265Dec struct {
	Element
}

func NewLibDe265Dec(name string) (*LibDe265Dec, error) {
	element, err := makeElement(name, "libde265dec")

	if err != nil {
		return nil, err
	}

	libDe265Dec := LibDe265Dec{
		element,
	}
	enableGarbageCollection(&libDe265Dec)

	return &libDe265Dec, nil
}

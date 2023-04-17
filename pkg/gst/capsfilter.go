package gst

type CapsFilter struct {
	Element
}

func NewCapsFilter(name string) (*CapsFilter, error) {
	element, err := makeElement(name, "capsfilter")

	if err != nil {
		return nil, err
	}

	capsFilter := CapsFilter{element}
	enableGarbageCollection(&capsFilter)

	return &capsFilter, nil
}

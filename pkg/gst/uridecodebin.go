package gst

type UriDecodeBin struct {
	Element
}

func NewUriDecodeBin(name string, uri string) (*UriDecodeBin, error) {
	element, err := makeElement(name, "uridecodebin")

	if err != nil {
		return nil, err
	}

	uriDecodeBin := UriDecodeBin{element}
	enableGarbageCollection(&uriDecodeBin)
	uriDecodeBin.SetProperty("uri", uri)

	return &uriDecodeBin, nil
}

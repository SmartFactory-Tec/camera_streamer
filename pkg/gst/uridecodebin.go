package gst

type UriDecodeBin struct {
	*BaseElement
}

func NewUriDecodeBin(name string, uri string) (UriDecodeBin, error) {
	createdElement, err := NewGstElement("uridecodebin", name)

	if err != nil {
		return UriDecodeBin{}, err
	}

	createdElement.SetProperty("uri", uri)

	return UriDecodeBin{&createdElement}, nil
}

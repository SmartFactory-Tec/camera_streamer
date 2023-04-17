package gst

type VideoScale struct {
	Element
}

func NewVideoScale(name string) (*VideoScale, error) {
	element, err := makeElement(name, "videoscale")

	if err != nil {
		return nil, err
	}

	videoScale := VideoScale{element}
	enableGarbageCollection(&videoScale)

	return &videoScale, nil
}

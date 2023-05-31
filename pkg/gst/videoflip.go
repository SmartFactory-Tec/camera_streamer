package gst

type VideoFlip struct {
	Element
}

type VideoOrientationMethod int

const (
	IDENTITY                   VideoOrientationMethod = 0
	ROTATE_CLOCKWISE_90        VideoOrientationMethod = 1
	ROTATE_180                 VideoOrientationMethod = 2
	ROTATE_COUNTERCLOCKWISE_90 VideoOrientationMethod = 3
)

func NewVideoFlip(name string, orientation VideoOrientationMethod) (*VideoFlip, error) {
	element, err := makeElement(name, "videoflip")

	if err != nil {
		return nil, err
	}

	videoFlip := VideoFlip{element}
	enableGarbageCollection(&videoFlip)

	if err := videoFlip.SetProperty("video-direction", int(orientation)); err != nil {
		return nil, err
	}

	return &videoFlip, nil
}

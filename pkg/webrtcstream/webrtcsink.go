package webrtcstream

import "C"
import (
	"context"
	"github.com/SmartFactory-Tec/camera_streamer/pkg/gst"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

type WebRtcSink struct {
	*gst.AppSink
	track *webrtc.TrackLocalStaticSample
}

func NewWebRtcSink(name string, track *webrtc.TrackLocalStaticSample) (*WebRtcSink, error) {
	createdAppSink, err := gst.NewAppSink(name)
	if err != nil {
		return nil, err
	}

	if err := createdAppSink.SetProperty("sync", true); err != nil {
		return nil, err
	}

	return &WebRtcSink{createdAppSink, track}, nil
}

func (w *WebRtcSink) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// If context is done stop loop
			return
		default:
			sample, err := w.PullSample()

			if err != nil {
				// Don't process if nothing is available yet
				continue
			}

			buffer := sample.Buffer()
			data := buffer.Bytes()
			duration := buffer.Duration()

			if err := w.track.WriteSample(media.Sample{
				Data:     data,
				Duration: duration,
			}); err != nil {
				break
			}
		}

	}

}

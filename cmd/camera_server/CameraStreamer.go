package main

import (
	"camera_server/pkg/signal"
	"go.uber.org/zap"
	"net/http"
)

// CameraStreamer represents the streaming infrastructure required to send video to the browser via webrtc.
type CameraStreamer struct {
	source   *Stream
	signaler *signal.Signaler
	logger   *zap.SugaredLogger
}

// NewCameraStreamer constructs a new streamer from a source stream and a given resolution, ready to be used in
// any net/http compliant web server
func NewCameraStreamer(source *Stream, logger *zap.SugaredLogger) CameraStreamer {
	streamLogger := logger.Named("CameraStreamer").With("streamName", source.Id)

	logger.Debugw("Obtaining track for target")
	sourceTrack, err := source.GetTrack()
	signaler := signal.NewSignaler(sourceTrack, logger)
	signaler.OnConnectionClosed(func() {
		localLogger := logger.Named("OnConnectionClosed")
		localLogger.Debugw("ending stream")
		if err := source.EndTrack(signaler.SourceTrack); err != nil {
			localLogger.Errorw(err.Error())
		}
	})
	if err != nil {
		logger.Errorw("unable to obtain track for requested resolution", "error", err.Error())
	}
	return CameraStreamer{
		source, &signaler, streamLogger,
	}
}

// Begin hijacks a http request and begins a signaling session, upgrading the connection to a websocket and
// beginning the webrtc connection.
func (c *CameraStreamer) Begin(w http.ResponseWriter, r *http.Request) {
	c.logger.Debugw("Starting signaling session")
	c.signaler.StartSignaling(w, r)

}

package main

import (
	"context"
	"github.com/SmartFactory-Tec/camera_streamer/pkg/webrtcstream"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"net/http"
)

var webrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func makeTrackHandler(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger) webrtcstream.TrackRequestHandler {
	logger = logger.Named("TrackHandler")
	return func(ctx context.Context, track webrtc.TrackLocal) {
		HandleWebRTC(w, r, []webrtc.TrackLocal{track}, logger)
	}
}

func makeGetStreamHandler(logger *zap.SugaredLogger) http.HandlerFunc {
	logger = logger.Named("GetStreamHandler")
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		stream := ctx.Value("stream").(*webrtcstream.WebRTCStream)

		logger.Info("starting webrtc session")

		stream.HandleTrackRequest(ctx, logger, makeTrackHandler(w, r, logger))

		logger.Info("webrtc session ended")
	}
}

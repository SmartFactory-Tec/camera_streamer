package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"net/http"
	"nhooyr.io/websocket"
)

var webrtcConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func WebRTCHandler(clientOrigin string, httpsOnly bool, allowAllOrigins bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*zap.SugaredLogger).Named("WebRTCHandler")
		sourceStream := ctx.Value("stream").(*Stream)

		sourceTrack, err := sourceStream.GenerateTrack()
		defer func() {
			err := sourceStream.StopTrack(sourceTrack)
			if err != nil {
				logger.Errorw("error stopping track", err)
			}
		}()
		if err != nil {
			logger.Errorw("error creating track", err)
		}
		context.WithValue(ctx, "track", sourceTrack)

		var origin string
		if httpsOnly {
			origin = fmt.Sprintf("https://%s", clientOrigin)
		} else {
			origin = clientOrigin
		}

		acceptOptions := websocket.AcceptOptions{
			OriginPatterns:     []string{origin},
			InsecureSkipVerify: allowAllOrigins,
		}

		socket, err := websocket.Accept(w, r, &acceptOptions)
		if err != nil {
			logger.Errorw("error opening socket", "error", err)
			return
		}
		defer func() {
			err := socket.Close(websocket.StatusInternalError, "internal error")
			if err != nil {
				logger.Panicw("unable to close socket", err)
			}
		}()

		// create RTCPeerConnection
		logger.Debugw("Creating peer connection")
		peerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
		if err != nil {
			logger.Errorw("Error establishing peer connection")
			err := peerConnection.Close()
			if err != nil {
				logger.Errorw("Error closing peer connection")
			}
		}

		_, err = peerConnection.AddTransceiverFromTrack(sourceTrack)
		if err != nil {
			logger.Errorw("Error adding transceiver from track", "err", err.Error())
			err := peerConnection.Close()
			if err != nil {
				logger.Errorw("Error closing peer connection")
			}
		}

		err = BeginSignalingSession(ctx, peerConnection, socket, logger)
		if err != nil {
			logger.Error(fmt.Errorf("signaling session failed: %w", err))
		}

		var errs error

		errs = multierror.Append(errs, socket.Close(websocket.StatusNormalClosure, "session finished normally"), peerConnection.Close())

		if errs != nil {
			logger.Errorw("error closing signaling session", "error", errs)
		}
	}
}

package main

import (
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
	if httpsOnly {
		clientOrigin = fmt.Sprintf("https://%s", clientOrigin)
	}
	acceptOptions := websocket.AcceptOptions{
		OriginPatterns:     []string{clientOrigin},
		InsecureSkipVerify: allowAllOrigins,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*zap.SugaredLogger).Named("WebRTCHandler")
		sourceStream := ctx.Value("stream").(*Stream)

		logger.Info("starting webrtc session")

		sourceTrack, err := sourceStream.GenerateTrack()
		if err != nil {
			logger.Error(fmt.Errorf("error creating track: %w", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			err := sourceStream.StopTrack(sourceTrack)
			if err != nil {
				logger.Error(fmt.Errorf("error cleaning up track: %w", err))
			}
		}()

		socket, err := websocket.Accept(w, r, &acceptOptions)
		if err != nil {
			logger.Error(fmt.Errorf("error opening socket: %w", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			// This *should* return an error, as the socket should be closed by the client before reaching this point.
			err := socket.Close(websocket.StatusInternalError, "internal error")
			if err == nil {
				logger.Error(fmt.Errorf("socket closed abruptly: %w", err))
			}
		}()

		// create RTCPeerConnection
		logger.Debugw("Creating peer connection")
		peerConnection, err := webrtc.NewPeerConnection(webrtcConfig)
		if err != nil {
			logger.Error(fmt.Errorf("error creating peer connection: %w", err))

			err := socket.Close(websocket.StatusInternalError, "peer connection error")

			if err != nil {
				logger.Error(err)
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			err = peerConnection.Close()
			if err != nil {
				logger.Error(fmt.Errorf("error closing peer connection: %w", err))
			}
		}()

		_, err = peerConnection.AddTransceiverFromTrack(sourceTrack)
		if err != nil {
			logger.Error(fmt.Errorf("error adding transceiver from track: %w", err))
			errs := multierror.Append(socket.Close(websocket.StatusInternalError, "peer connection error"),
				peerConnection.Close())
			if errs.Len() != 0 {
				logger.Error(errs)
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = BeginSignalingSession(ctx, peerConnection, socket, logger)
		if err != nil {
			logger.Error(fmt.Errorf("signaling session failed: %w", err))
			w.WriteHeader(http.StatusInternalServerError)
		}

		logger.Info("webrtc session ended")
	}
}

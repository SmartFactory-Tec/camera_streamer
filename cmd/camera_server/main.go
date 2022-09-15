package main

import (
	"camera_server/pkg/gst"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"net/http"
)

// TODO the gst library is most likely full of memory leaks, fix

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO add parameters for this
		return true
		//return r.Header.Get("Origin") == "http://localhost:3001"
	},
}

func main() {
	baseLogger, _ := zap.NewDevelopment()
	logger := baseLogger.Sugar().Named("main")

	logger.Debugw("Initializing GStreamer")
	gst.Init()

	camera := NewCameraStream("cam2", "10.22.240.53", 554, "cam/realmonitor?channel=1&subtype=0&proto=Onvif", "admin", "L2793C70", logger)
	end := make(chan bool)
	go camera.StartMsgBus(end, logger)

	getStream := func(w http.ResponseWriter, r *http.Request) {
		logger := logger.Named("getStream")
		logger.Infow("WebRTC stream requested, upgrading to websocket")
		// Upgrade http connection to a websocket
		socket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorw("Unable to upgrade connection to websocket")
			return
		}

		// create RTCPeerConnection
		logger.Debugw("Creating peer connection")
		peerConnection, err := webrtc.NewPeerConnection(config)
		panicIfError(err)

		// create video track
		videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType: "video/vp8",
		}, fmt.Sprintf("cam2-%s", uuid.New()), "mainStream")

		_, err = peerConnection.AddTrack(videoTrack)
		panicIfError(err)

		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			logger := logger.Named("OnICEConnectionStateChange").With("receivedIceState", connectionState.String())
			logger.Debugw("Connection state changed")
			if connectionState == webrtc.ICEConnectionStateConnected {
				logger.Debugw("Adding output track", "track", videoTrack.ID())
				println("Adding output track")
				camera.AddOutputTrack(videoTrack, logger)
			} else if connectionState == webrtc.ICEConnectionStateDisconnected {
				logger.Debugw("Removing output track", "track", videoTrack.ID())
				err := camera.RemoveOutputTrack(videoTrack, logger)
				if err != nil {
					logger.Errorw("Error removing output track", "err", err.Error())

				}
				logger.Infow("Closing connection to peer")
				err = peerConnection.Close()
				if err != nil {
					logger.Panicw("Error closing peerConnection", "err", err.Error())
				}
			}
		})

		// process incoming offer
		offer := webrtc.SessionDescription{}
		logger.Debugw("Receiving remote description")
		err = socket.ReadJSON(&offer)
		if err != nil {
			logger.Errorw("Error reading remote description from peer", "err", err.Error())
			err := socket.Close()
			panicIfError(err)
			return
		}

		err = peerConnection.SetRemoteDescription(offer)
		if err != nil {
			logger.Errorw("Error processing remote description", "err", err.Error())
			err := socket.Close()
			panicIfError(err)
			return
		}

		//create anwser
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			logger.Errorw("Failed to create answer for peer", "err", err.Error())
			err := socket.Close()
			panicIfError(err)
			return
		}
		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			logger.Errorw("Setting local description from answer failed", "err", err.Error())
			err := socket.Close()
			panicIfError(err)
			return
		}

		<-gatherComplete

		err = socket.WriteJSON(peerConnection.LocalDescription())
		if err != nil {
			logger.Errorw("Failed to send local description to peer", "err", err.Error())
			err := socket.Close()
			panicIfError(err)
			return
		}

		logger.Debugw("Closing websocket connection")
		err = socket.Close()
		panicIfError(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream", getStream)

	handler := cors.Default().Handler(mux)

	err := http.ListenAndServe(":3000", handler)

	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)

	}

	end <- true
}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

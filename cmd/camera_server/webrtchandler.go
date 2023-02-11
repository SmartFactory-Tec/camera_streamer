// Package signal encapsulates all code required to initiate a signaling session with a client and send video to it.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WebRTCConfig struct {
	AllowAllOrigins bool
	ClientOrigins   []string
}

// HandleWebRTC configures the signaling session utilizing a given context
func HandleWebRTC(w http.ResponseWriter, r *http.Request, tracks []webrtc.TrackLocal, logger *zap.SugaredLogger) {
	logger = logger.Named("HandleWebRTC")

	acceptOptions := websocket.AcceptOptions{
		InsecureSkipVerify: true,
	}

	logger.Debugw("opening websocket")

	// Open socket for signaling session
	socket, err := websocket.Accept(w, r, &acceptOptions)
	if err != nil {
		logger.Error(fmt.Errorf("error opening socket: %w", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Debugw("opening peer connection")
	// Open peer connection
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

	// Create a context for the signaling session
	ctx := r.Context()
	signalingCtx, cancelSignaling := context.WithCancel(r.Context())

	// Set up peer connection callbacks
	peerConnection.OnNegotiationNeeded(makeNegotiationNeededHandler(signalingCtx, peerConnection, socket, logger))
	peerConnection.OnICECandidate(makeIceCandidateHandler(ctx, socket, logger))
	peerConnection.OnSignalingStateChange(makeSignalingStateChangeHandler(cancelSignaling, logger))
	peerConnection.OnICEConnectionStateChange(makeIceConnectionStateHandler(cancelSignaling, logger))
	peerConnection.OnICEGatheringStateChange(makeIceGatheringStateChangeHandler(logger))

	for _, track := range tracks {
		logger.Debugw("adding track to peer connection", "track id", track.ID(), "stream id", track.StreamID())
		_, err := peerConnection.AddTransceiverFromTrack(track)
		if err != nil {
			logger.Errorw("could not add track", "track id", track.ID(), "stream id", track.StreamID())
		}
	}

	HandleSignalingSession(signalingCtx, socket, peerConnection, logger)
}

func HandleSignalingSession(ctx context.Context, socket *websocket.Conn, peerConnection *webrtc.PeerConnection, logger *zap.SugaredLogger) {
	logger = logger.Named("HandleSignalingSession")

	for {
		select {
		// If parent context has been canceled
		case <-ctx.Done():
			return
		default:
		}

		// Blocks until peer sends a message
		logger.Debugw("awaiting message from socket")
		message := Message{}
		err := wsjson.Read(ctx, socket, &message)
		var closeError websocket.CloseError
		if errors.As(err, &closeError) {
			switch closeError.Code {
			case websocket.StatusNormalClosure:
				fallthrough
			case websocket.StatusGoingAway:
				logger.Debugw("socket closed", "reason", closeError.Code.String())
				return
			default:
				err = fmt.Errorf("socket closed: %w", err)
				logger.Error(err)
				return
			}
		} else if err != nil {
			err = fmt.Errorf("error reading from socket: %w", err)
			logger.Error(err)
			return
		}
		logger.Debugw("got message from socket", "type", message.MsgType)

		switch message.MsgType {
		case SESSION_DESCRIPTION:
			sessionDescription, err := message.SessionDescription()
			if err != nil {
				logger.Error(fmt.Errorf("error parsing session description: %w", err))
				return
			}

			handleSessionDescription(ctx, sessionDescription, peerConnection, socket, logger)
		case ICE_CANDIDATE:
			iceCandidate, err := message.IceCandidate()
			if err != nil {
				logger.Error(fmt.Errorf("error parsing ice candidate: %w", err))
			}

			handleIceCandidate(iceCandidate, peerConnection, logger)
		default:
			logger.Errorw("unknown message type received from peer", "message type", message.MsgType)
		}
	}
}

func makeSignalingStateChangeHandler(cancelSignaling context.CancelFunc, logger *zap.SugaredLogger) func(state webrtc.SignalingState) {
	logger = logger.Named("SignalingStateChangeHandler")
	return func(state webrtc.SignalingState) {
		logger.Debugw("signaling state has changed", "new state", state.String())
		switch state {
		case webrtc.SignalingStateClosed:
			cancelSignaling()
		}
	}
}

func makeIceCandidateHandler(ctx context.Context, socket *websocket.Conn, logger *zap.SugaredLogger) func(candidate *webrtc.ICECandidate) {
	logger = logger.Named("IceCandidateHandler")
	return func(candidate *webrtc.ICECandidate) {
		message := Message{
			MsgType: ICE_CANDIDATE,
		}

		if candidate == nil {
			logger.Debugw("end of candidates")
			message.Payload = nil
		} else {
			logger.Debug("new ice candidate")
			message.Payload = candidate.ToJSON()
		}

		logger.Debug("sending candidate to peer")
		err := wsjson.Write(ctx, socket, message)

		if err != nil {
			logger.Error(fmt.Errorf("could not send ice candidate to peer: %w", err))
			return
		}
	}
}

func makeNegotiationNeededHandler(ctx context.Context, peerConnection *webrtc.PeerConnection, socket *websocket.Conn, logger *zap.SugaredLogger) func() {
	logger = logger.Named("NegotiationNeededHandler")
	return func() {
		logger.Debugw("starting negotiation")

		offerOptions := webrtc.OfferOptions{}

		logger.Debugw("creating offer")
		offer, err := peerConnection.CreateOffer(&offerOptions)
		if err != nil {
			logger.Error(fmt.Errorf("error creating offer: %w", err))
			return
		}
		logger.Debugw("setting local description from offer")
		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			logger.Error(fmt.Errorf("error setting local description from new offer: %w", err))
			return
		}
		message := Message{MsgType: SESSION_DESCRIPTION, Payload: offer}

		logger.Debugw("sending local description to peer")

		err = wsjson.Write(ctx, socket, message)
		if err != nil {
			logger.Error(fmt.Errorf("error sending local description to peer: %w", err))
			return
		}
	}
}

func makeIceConnectionStateHandler(cancelSignaling context.CancelFunc, logger *zap.SugaredLogger) func(state webrtc.ICEConnectionState) {
	logger = logger.Named("IceConnectionStateHandler")
	return func(state webrtc.ICEConnectionState) {
		logger.Debugw("ice connection state changed", "state", state.String())

		switch state {
		case webrtc.ICEConnectionStateFailed:
			logger.Error("failed to connect to peer")
			cancelSignaling()
		case webrtc.ICEConnectionStateClosed:
			logger.Debugw("connection closed")
			cancelSignaling()
		}
	}
}

func makeIceGatheringStateChangeHandler(logger *zap.SugaredLogger) func(state webrtc.ICEGathererState) {
	logger = logger.Named("IceGatheringStateChangeHandler")
	return func(state webrtc.ICEGathererState) {
		logger.Debugw("ice gathering state changed", "state", state.String())
	}
}

func handleSessionDescription(ctx context.Context, sessionDescription webrtc.SessionDescription, peerConnection *webrtc.PeerConnection, socket *websocket.Conn, logger *zap.SugaredLogger) {
	logger = logger.Named("handleSessionDescription")
	logger.Debugw("received session description")
	switch peerConnection.SignalingState() {
	case webrtc.SignalingStateHaveLocalOffer:
		logger.Debugw("have local offer, setting remote sessionDescription")
		err := peerConnection.SetRemoteDescription(sessionDescription)
		if err != nil {
			// TODO handle error
			logger.Error(fmt.Errorf("error setting remote description: %w", err))
			return
		}
	case webrtc.SignalingStateStable:
		logger.Debugw("setting remote sessionDescription")
		err := peerConnection.SetRemoteDescription(sessionDescription)

		if err != nil {
			logger.Error(fmt.Errorf("error setting remote description: %w", err))
			return
		}

		logger.Debugw("creating answer")
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			logger.Error(fmt.Errorf("could not create answer: %w", err))
			return
		}

		logger.Debugw("setting local sessionDescription from answer")
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			logger.Error(fmt.Errorf("could not set local description from answer: %w", err))
			return
		}

		logger.Debugw("sending answer to peer")
		message := Message{MsgType: SESSION_DESCRIPTION, Payload: answer}
		err = wsjson.Write(ctx, socket, message)
		if err != nil {
			logger.Error(fmt.Errorf("error sending answer to peer: %w", err))
			return
		}
	}
}

func handleIceCandidate(iceCandidate webrtc.ICECandidateInit, peerConnection *webrtc.PeerConnection, logger *zap.SugaredLogger) {
	logger = logger.Named("handleIceCandidate")
	logger.Debugw("adding ice candidate")
	err := peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		logger.Error(fmt.Errorf("error adding ice candidate: %w", err))
	}
}

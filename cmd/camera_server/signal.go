// Package signal encapsulates all code required to initiate a signaling session with a client and send video to it.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// BeginSignalingSession configures the signaling session utilizing a given context
func BeginSignalingSession(ctx context.Context, peerConnection *webrtc.PeerConnection, socket *websocket.Conn, logger *zap.SugaredLogger) error {
	logger = logger.Named("BeginSignalingSession")

	ctx, cancel := context.WithCancelCause(ctx)

	// Set up peer connection callbacks and erase them once signaling session ends
	peerConnection.OnNegotiationNeeded(NegotiationNeededHandler(ctx, cancel, peerConnection, socket, logger))
	defer peerConnection.OnNegotiationNeeded(nil)

	peerConnection.OnICECandidate(ICECandidateHandler(ctx, socket, logger))
	defer peerConnection.OnICECandidate(nil)

	peerConnection.OnSignalingStateChange(SignalingStateChangeHandler(logger))
	defer peerConnection.OnSignalingStateChange(nil)

	peerConnection.OnICEConnectionStateChange(ICEConnectionStateHandler(cancel, logger))
	defer peerConnection.OnICEConnectionStateChange(nil)

	peerConnection.OnICEGatheringStateChange(ICEGatherinStateHandler(logger))
	defer peerConnection.OnICEGatheringStateChange(nil)

	// Start signaling loop
	for {
		select {
		// If parent context has been canceled
		case <-ctx.Done():
			return context.Cause(ctx)
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
				return nil
			default:
				err = fmt.Errorf("socket closed: %w", err)
				logger.Error(err)
				return err
			}
		} else if err != nil {
			err = fmt.Errorf("error reading from socket: %w", err)
			logger.Error(err)
			return err
		}
		logger.Debugw("got message from socket", "type", message.MsgType)

		switch message.MsgType {
		case SESSION_DESCRIPTION:
			sessionDescription, err := message.SessionDescription()

			if err != nil {
				return fmt.Errorf("error parsing session description: %w", err)
			}

			err = onSessionDescription(ctx, sessionDescription, peerConnection, socket, logger)
			if err != nil {
				return err
			}
		case ICE_CANDIDATE:
			iceCandidate, err := message.IceCandidate()

			if err != nil {
				return fmt.Errorf("error parsing ice candidate: %w", err)
			}

			err = onICECandidate(iceCandidate, peerConnection)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid message received of type %d", message.MsgType)
		}
	}
}

func onSessionDescription(ctx context.Context, sessionDescription webrtc.SessionDescription, peerConnection *webrtc.PeerConnection, socket *websocket.Conn, logger *zap.SugaredLogger) error {
	logger = logger.Named("onSessionDescription")
	logger.Debugw("Received session sessionDescription from peer")
	switch peerConnection.SignalingState() {
	case webrtc.SignalingStateHaveLocalOffer:
		logger.Debugw("Have local offer, receiving answer and setting remote sessionDescription")
		err := peerConnection.SetRemoteDescription(sessionDescription)
		if err != nil {
			return fmt.Errorf("error setting remote description: %w", err)
		}
	case webrtc.SignalingStateStable:
		logger.Debugw("Receiving offer and setting remote sessionDescription")
		err := peerConnection.SetRemoteDescription(sessionDescription)

		if err != nil {
			return fmt.Errorf("error setting remote description: %w", err)
		}

		logger.Debugw("Creating answer")
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			return fmt.Errorf("could not create answer: %w", err)
		}

		logger.Debugw("Setting local sessionDescription")
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			return fmt.Errorf("could not set local description from answer: %w, err")
		}

		logger.Debugw("Sending answer to peer")
		message := Message{MsgType: SESSION_DESCRIPTION, Payload: answer}
		err = wsjson.Write(ctx, socket, message)
		if err != nil {
			return fmt.Errorf("error sending answer to peer: %w", err)
		}
	}
	return nil
}

func onICECandidate(iceCandidate webrtc.ICECandidateInit, peerConnection *webrtc.PeerConnection) error {
	err := peerConnection.AddICECandidate(iceCandidate)
	return err
}

func SignalingStateChangeHandler(logger *zap.SugaredLogger) func(state webrtc.SignalingState) {
	logger = logger.Named("SignalingStateChangeHandler")
	return func(state webrtc.SignalingState) {
		logger.Debugw("signaling state has changed", "new state", state.String())
	}
}

func ICECandidateHandler(ctx context.Context, socket *websocket.Conn, logger *zap.SugaredLogger) func(candidate *webrtc.ICECandidate) {
	logger = logger.Named("ICECandidateHandler")
	return func(candidate *webrtc.ICECandidate) {
		message := Message{
			MsgType: ICE_CANDIDATE,
		}

		if candidate == nil {
			logger.Debugw("end of candidates")
			message.Payload = nil
		} else {
			logger.Debug("New ICE Candidate")
			message.Payload = candidate.ToJSON()
		}

		err := wsjson.Write(ctx, socket, message)

		if err != nil {
			logger.Errorw("could not send ice candidate to peer", "err", err)
		}
	}
}

func NegotiationNeededHandler(ctx context.Context, cancel context.CancelCauseFunc, peerConnection *webrtc.PeerConnection, socket *websocket.Conn, logger *zap.SugaredLogger) func() {
	return func() {
		logger.Debugw("starting negotiation")

		offerOptions := webrtc.OfferOptions{}

		logger.Debugw("creating offer")
		offer, err := peerConnection.CreateOffer(&offerOptions)
		if err != nil {
			logger.Errorw("error creating offer")
			cancel(fmt.Errorf("error creating offer: %w", err))
			return
		}
		logger.Debugw("setting local description from offer")
		err = peerConnection.SetLocalDescription(offer)
		if err != nil {
			logger.Errorw("error setting local description from new offer")
			cancel(fmt.Errorf("could not set local description: %w", err))
			return
		}
		message := Message{MsgType: SESSION_DESCRIPTION, Payload: offer}

		logger.Debugw("sending local description to peer")

		err = wsjson.Write(ctx, socket, message)
		if err != nil {
			logger.Errorw("error sending local description to peer")
			cancel(fmt.Errorf("could not send local description to peer: %w", err))
			return
		}
	}
}

func ICEConnectionStateHandler(cancel context.CancelCauseFunc, logger *zap.SugaredLogger) func(state webrtc.ICEConnectionState) {
	return func(state webrtc.ICEConnectionState) {
		logger.Debugw("ice connection state changed")

		switch state {
		case webrtc.ICEConnectionStateConnected:
		case webrtc.ICEConnectionStateDisconnected:
			cancel(nil)
		case webrtc.ICEConnectionStateFailed:
		case webrtc.ICEConnectionStateClosed:
		}
	}
}

func ICEGatherinStateHandler(logger *zap.SugaredLogger) func(state webrtc.ICEGathererState) {
	return func(state webrtc.ICEGathererState) {
		logger.Debugw("ice gathering state changed")
	}
}

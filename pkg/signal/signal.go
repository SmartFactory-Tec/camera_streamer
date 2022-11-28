// Package signal encapsulates all code required to initiate a signaling session with a client and send video to it.
package signal

import (
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// Signaler implements a WebRTC signaling server via a websocket connection. It offers a transceiver for the
// source track it was given.
type Signaler struct {
	SourceTrack              *webrtc.TrackLocalStaticSample
	upgrader                 websocket.Upgrader
	socket                   *websocket.Conn
	socketMutex              sync.Mutex
	logger                   *zap.SugaredLogger
	peerConnection           *webrtc.PeerConnection
	closedConnectionCallback func()
	closed                   bool
}

// NewSignaler creates a signaler with a given source track
func NewSignaler(sourceTrack *webrtc.TrackLocalStaticSample, logger *zap.SugaredLogger) Signaler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO add parameters for this
			return true
			//return r.Header.Get("Origin") == "http://localhost:3001"
		},
	}

	return Signaler{
		sourceTrack,
		upgrader,
		nil,
		sync.Mutex{},
		logger.Named("Signaler"),
		nil,
		func() {},
		false,
	}
}

// StartSignaling hijacks a http request, upgrading it to a websocket connection and beginning the messaging loop
func (s *Signaler) StartSignaling(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.Named("StartSignaling")
	logger.Infow("Ugrading connection to websocket")

	// Upgrade http connection to a websocket
	var err error
	s.socket, err = s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorw("Unable to upgrade connection to websocket")
		return
	}

	// create RTCPeerConnection
	logger.Debugw("Creating peer connection")
	s.peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		logger.Errorw("Error establishing peer connection")
		s.CloseSession()
		return
	}

	// Set up peer connection callbacks
	s.peerConnection.OnNegotiationNeeded(s.onNegotiationNeeded)

	s.peerConnection.OnICECandidate(s.onICECandidate)

	s.peerConnection.OnSignalingStateChange(s.onSignalingStateChange)

	s.peerConnection.OnICEConnectionStateChange(s.onICEConnectionStateChange)

	s.peerConnection.OnICEGatheringStateChange(s.onICEGatheringStateChange)

	_, err = s.peerConnection.AddTransceiverFromTrack(s.SourceTrack)
	if err != nil {
		logger.Errorw("Error adding transceiver from track", "err", err.Error())
		s.CloseSession()
	}

	// Start signaling loop
	s.processSignals()
}

// processSignals is the signaling loop. It captures incoming websocket messages, and processes them into the webrtc client.
func (s *Signaler) processSignals() {
	for {
		if s.closed {
			break
		}

		logger := s.logger.Named("processSignals")

		message := Message{}
		err := s.socket.ReadJSON(&message)
		// Receive message
		logger.Debugw("Received message from peer")
		if err != nil {
			logger.Errorw(err.Error())
			s.CloseSession()
			break
		}

		// If message is a session description
		sessionDescription, err := message.SessionDescription()
		if err == nil {
			err := s.onSessionDescriptionSignal(sessionDescription)
			if err != nil {
				s.logger.Errorw("error processing session description signal", "err", err.Error())
				s.CloseSession()
			}
			continue
		}

		// if message is an ice candidate
		iceCandidate, err := message.IceCandidate()
		if err == nil {
			err := s.onICECandidateSignal(iceCandidate)
			if err != nil {
				s.logger.Errorw("error processing ice candidate signal", "err", err.Error())
				s.CloseSession()
			}
			continue
		}

		// else
		logger.Errorw("Invalid message received from peer", "MsgType", message.MsgType)
	}
}

func (s *Signaler) onSessionDescriptionSignal(sessionDescription webrtc.SessionDescription) error {
	logger := s.logger.Named("onSessionDescriptionSignal")
	logger.Debugw("Received session sessionDescription from peer")
	switch s.peerConnection.SignalingState() {
	case webrtc.SignalingStateHaveLocalOffer:
		logger.Debugw("Have local offer, receiving answer and setting remote sessionDescription")
		err := s.peerConnection.SetRemoteDescription(sessionDescription)
		if err != nil {
			return err
		}
	case webrtc.SignalingStateStable:
		logger.Debugw("Receiving offer and setting remote sessionDescription")
		err := s.peerConnection.SetRemoteDescription(sessionDescription)

		if err != nil {
			return err
		}

		logger.Debugw("Creating answer")
		answer, err := s.peerConnection.CreateAnswer(nil)
		if err != nil {
			return err
		}

		logger.Debugw("Setting local sessionDescription")
		err = s.peerConnection.SetLocalDescription(answer)
		if err != nil {
			return err
		}

		logger.Debugw("Sending answer to peer")
		message := Message{MsgType: SESSION_DESCRIPTION, Payload: answer}
		err = s.socket.WriteJSON(message)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *Signaler) onICECandidateSignal(iceCandidate webrtc.ICECandidateInit) error {
	err := s.peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		s.CloseSession()
		return err
	}
	return nil
}

func (s *Signaler) CloseSession() {
	if s.closed {
		return
	}

	s.closed = true

	logger := s.logger.Named("CloseSession")

	logger.Debugw("Closing callback")
	s.closedConnectionCallback()

	logger.Debugw("Closing peer connection")
	if err := s.peerConnection.Close(); err != nil {
		logger.Error(err.Error())
	}

	logger.Infow("Closing signaling socket")

	if err := s.socket.Close(); err != nil {
		logger.DPanic("Unable to close websocket ")
	}

}

func (s *Signaler) onSignalingStateChange(state webrtc.SignalingState) {

	logger := s.logger.Named("onSignalingStateChange").With("newSignalingState", state.String())
	logger.Debugw("Signaling state changed")
	if state == webrtc.SignalingStateHaveLocalPranswer {

	}
}

func (s *Signaler) onICECandidate(candidate *webrtc.ICECandidate) {
	logger := s.logger.Named("onICECandidate")

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

	s.socketMutex.Lock()
	defer s.socketMutex.Unlock()

	err := s.socket.WriteJSON(message)

	if err != nil {
		s.CloseSession()
	}
}

func (s *Signaler) onNegotiationNeeded() {
	logger := s.logger.Named("onNegotiationNeeded")
	logger.Debugw("Starting negotiation")

	offerOptions := webrtc.OfferOptions{}

	logger.Debugw("Creating offer")
	offer, err := s.peerConnection.CreateOffer(&offerOptions)
	if err != nil {
		logger.Errorw("Error creating offer")
		s.CloseSession()
		return
	}
	logger.Debugw("Setting local description from offer")
	err = s.peerConnection.SetLocalDescription(offer)
	if err != nil {
		logger.Errorw("Error setting local description from new offer")
		s.CloseSession()
		return
	}
	message := Message{MsgType: SESSION_DESCRIPTION, Payload: offer}

	s.socketMutex.Lock()
	defer s.socketMutex.Unlock()

	logger.Debugw("Sending local description to peer")
	err = s.socket.WriteJSON(message)

	if err != nil {
		logger.Errorw("Error sending local description to peer")
	}

}

func (s *Signaler) onICEConnectionStateChange(state webrtc.ICEConnectionState) {
	logger := s.logger.Named("onICEConnectionStateChange").With("newIceState", state.String())
	logger.Debugw("ICE connection state changed")

	switch state {
	case webrtc.ICEConnectionStateConnected:

	case webrtc.ICEConnectionStateDisconnected:
		s.CloseSession()
	case webrtc.ICEConnectionStateFailed:
	case webrtc.ICEConnectionStateClosed:
	}
}

func (s *Signaler) onICEGatheringStateChange(state webrtc.ICEGathererState) {
	logger := s.logger.Named("onICEGatheringStateChange").With("newGatheringState", state)
	logger.Debugw("ICE gathering state changed")
}

func (s *Signaler) OnConnectionClosed(callback func()) {
	s.closedConnectionCallback = callback
}

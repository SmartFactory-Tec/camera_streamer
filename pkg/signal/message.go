package signal

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/webrtc/v3"
)

type MsgType int

const (
	SESSION_DESCRIPTION MsgType = iota
	ICE_CANDIDATE
	STREAMS_DESCRIPTION
)

type Message struct {
	MsgType MsgType `json:"type"`
	Payload any     `json:"payload"`
}

func (m Message) String() string {
	switch m.MsgType {
	case SESSION_DESCRIPTION:
		return "Session description"
	case ICE_CANDIDATE:
		return "ICE Candidate"
	case STREAMS_DESCRIPTION:
		return "STREAM DESCRIPTION"
	default:
		return "UNKNOWN MESSAGE TYPE"
	}
}

type StreamDescription struct {
	ID   string `json:"id"`
	Size int    `json:"size"`
}

type StreamsDescription struct {
	Streams []StreamDescription `json:"streams"`
}

func getKeyAndCast[T any](inputMap map[string]interface{}, key string) (T, error) {
	v, ok := inputMap[key]
	if !ok {
		return *new(T), fmt.Errorf("unable to find value with key %s")
	}

	value, ok := v.(T)

	if !ok {
		return *new(T), fmt.Errorf("value with key %s is not of type %T", key, *new(T))
	}

	return value, nil
}

func (m *Message) IceCandidate() (webrtc.ICECandidateInit, error) {
	if m.MsgType != ICE_CANDIDATE {
		return webrtc.ICECandidateInit{}, fmt.Errorf("message is not an ICE candidate")
	}

	iceCandidate := webrtc.ICECandidateInit{}

	err := mapstructure.Decode(m.Payload, &iceCandidate)

	if err != nil {
		return webrtc.ICECandidateInit{}, err
	}

	return iceCandidate, nil
}

func (m *Message) SessionDescription() (webrtc.SessionDescription, error) {
	if m.MsgType != SESSION_DESCRIPTION {
		return webrtc.SessionDescription{}, fmt.Errorf("message is not a session description")
	}

	payload, ok := m.Payload.(map[string]interface{})
	if !ok {
		return webrtc.SessionDescription{}, fmt.Errorf("malformed payload for session description")
	}

	sdpTypeString, err := getKeyAndCast[string](payload, "type")
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	sdpType := webrtc.NewSDPType(sdpTypeString)

	sdp, err := getKeyAndCast[string](payload, "sdp")
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	return webrtc.SessionDescription{Type: sdpType, SDP: sdp}, nil
}

func (m *Message) StreamsDescription() (StreamsDescription, error) {
	if m.MsgType != STREAMS_DESCRIPTION {
		return StreamsDescription{}, fmt.Errorf("message is not a streams description")
	}

	streamsDescription := StreamsDescription{}

	err := mapstructure.WeakDecode(m.Payload, &streamsDescription)

	if err != nil {
		return StreamsDescription{}, err
	}

	return streamsDescription, nil
}

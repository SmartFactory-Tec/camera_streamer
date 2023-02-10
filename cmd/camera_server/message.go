package main

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/webrtc/v3"
	"reflect"
)

type MsgType int

const (
	SESSION_DESCRIPTION MsgType = iota
	ICE_CANDIDATE
	//STREAMS_DESCRIPTION
)

var PayloadParseError = errors.New("error parsing payload")

type Message struct {
	MsgType MsgType `json:"type"`
	Payload any     `json:"payload"`
}

//type StreamDescription struct {
//	ID   string `json:"id"`
//	Size int    `json:"size"`
//}

//type StreamsDescription struct {
//	Streams []StreamDescription `json:"streams"`
//}

//func getKeyAndCast[T any](inputMap map[string]interface{}, key string) (T, error) {
//	v, ok := inputMap[key]
//	if !ok {
//		return *new(T), fmt.Errorf("unable to find value with key %s")
//	}
//
//	value, ok := v.(T)
//
//	if !ok {
//		return *new(T), fmt.Errorf("value with key %s is not of type %T", key, *new(T))
//	}
//
//	return value, nil
//}

func (m Message) IceCandidate() (webrtc.ICECandidateInit, error) {
	if m.MsgType != ICE_CANDIDATE {
		return webrtc.ICECandidateInit{}, fmt.Errorf("message is not an ICE candidate")
	}

	iceCandidate := webrtc.ICECandidateInit{}

	err := mapstructure.Decode(m.Payload, &iceCandidate)

	if err != nil {
		return webrtc.ICECandidateInit{}, errors.Join(PayloadParseError, err)
	}

	return iceCandidate, nil
}

func (m Message) SessionDescription() (webrtc.SessionDescription, error) {
	if m.MsgType != SESSION_DESCRIPTION {
		return webrtc.SessionDescription{}, fmt.Errorf("message is not a session description")
	}

	sessionDescription := webrtc.SessionDescription{}

	config := mapstructure.DecoderConfig{
		Result: &sessionDescription,
		// ALlow conversion to SDPType
		DecodeHook: mapstructure.DecodeHookFuncType(func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
			if from.Kind() != reflect.String {
				return data, nil
			} else if to != reflect.TypeOf(webrtc.SDPType(0)) {
				return data, nil
			}
			return webrtc.NewSDPType(data.(string)), nil
		}),
	}

	decoder, err := mapstructure.NewDecoder(&config)

	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	err = decoder.Decode(m.Payload)

	if err != nil {
		return webrtc.SessionDescription{}, errors.Join(PayloadParseError, err)
	}

	return sessionDescription, nil
}

//func (m Message) StreamsDescription() (StreamsDescription, error) {
//	if m.MsgType != STREAMS_DESCRIPTION {
//		return StreamsDescription{}, fmt.Errorf("message is not a streams description")
//	}
//
//	streamsDescription := StreamsDescription{}
//
//	err := mapstructure.WeakDecode(m.Payload, &streamsDescription)
//
//	if err != nil {
//		return StreamsDescription{}, err
//	}
//
//	return streamsDescription, nil
//}

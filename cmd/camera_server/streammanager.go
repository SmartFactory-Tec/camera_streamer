package main

import "camera_server/pkg/gst"

func NewStreamStoreFromConfigs(streamConfigs []StreamConfig) map[string]*Stream {
	streamStore := make(map[string]*Stream)

	for _, streamConfig := range streamConfigs {
		var (
			streamSrc *gst.UriDecodeBin
			err       error
		)

		// Check if stream has auth specified
		if streamConfig.User != "" && streamConfig.Password != "" {
			streamSrc, err = NewUriDecodeBinWithAuth(streamConfig.Id, streamConfig.Hostname, streamConfig.Port, streamConfig.Path, streamConfig.User, streamConfig.Password)
		} else {
			streamSrc, err = NewUriDecodeBin(streamConfig.Id, streamConfig.Hostname, streamConfig.Port, streamConfig.Path)
		}
		if err != nil {
			panic(err)
		}

		stream, err := NewStream(streamConfig.Name, streamConfig.Id, streamSrc)
		if err != nil {
			panic(err)
		}

		streamStore[streamConfig.Id] = &stream
	}

	return streamStore
}

package main

import (
	"camera_server/pkg/gst"
	"fmt"
)

func NewUriDecodeBinWithAuth(name string, hostname string, port int, path string, user string, password string) (*gst.UriDecodeBin, error) {
	newBin, err := gst.NewUriDecodeBin(name, fmt.Sprintf("rtsp://%s:%s@%s:%d%s", user, password, hostname, port, path))

	if err != nil {
		return nil, err
	}

	return newBin, nil
}

func NewUriDecodeBin(name string, hostname string, port int, path string) (*gst.UriDecodeBin, error) {
	newBin, err := gst.NewUriDecodeBin(name, fmt.Sprintf("rtsp://%s:%d%s", hostname, port, path))

	if err != nil {
		return nil, err
	}

	return newBin, nil
}

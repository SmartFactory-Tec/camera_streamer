package main

import (
	"camera_server/pkg/gst"
	"fmt"
)

func NewUriDecodeBinWithAuth(name string, hostname string, port int, path string, user string, password string) (gst.UriDecodeBin, error) {
	newBin, err := gst.NewUriDecodeBin(name, fmt.Sprintf("rtsp://%s:%s@%s:%d%s", user, password, hostname, port, path))

	if err != nil {
		return gst.UriDecodeBin{}, err
	}

	return newBin, nil

	//camBin, err := gst.NewBaseBin(fmt.Sprintf("%s-bin", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//camera, err := gst.NewRtspSource(fmt.Sprintf("%s-rtspsrc", Id), fmt.Sprintf("rtsp://%s:%s@%s:%d%s", user, password, hostname, port, path))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//camera.SetProperty("udp-buffer-size", 212992)
	//
	//queue, err := gst.NewQueue(fmt.Sprintf("%s-queue", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//
	//depay, err := gst.NewRtpH265Depay(fmt.Sprintf("%s-depay", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//
	//parse, err := gst.NewH265Parse(fmt.Sprintf("%s-parse", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//
	//decode, err := gst.NewAvDecH265(fmt.Sprintf("%s-decode", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//}
	//
	//encode, err := gst.NewX264Enc(fmt.Sprintf("%s-encode", Id))
	//if err != nil {
	//	return RtspH265Cam{}, err
	//
	//if ok := camBin.AddElement(camera); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add rtsp source to bin")
	//}
	//if ok := camBin.AddElement(queue); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add queue to bin")
	//}
	//if ok := camBin.AddElement(depay); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add depay to bin")
	//}
	//if ok := camBin.AddElement(parse); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add parse to bin")
	//}
	//if ok := camBin.AddSrcElement(decode); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add decode to bin")
	//}
	//if ok := camBin.AddSrcElement(encode); !ok {
	//	return RtspH265Cam{}, fmt.Errorf("could not add encode to bin")
	//}
	//
	//if err = gst.LinkElements(queue, depay); err != nil {
	//	return RtspH265Cam{}, err
	//}
	//if err = gst.LinkElements(depay, decode); err != nil {
	//	return RtspH265Cam{}, err
	//}
	//gst.LinkElements(decode, encode)
	//if err = gst.LinkElements(decode, encode); err != nil {
	//	return RtspH265Cam{}, err
	//}

	//camera.OnPadAdded(func(newPad gst.Pad) {

	//})

	//return RtspH265Cam{
	//	&camBin,
	//}, nil
}

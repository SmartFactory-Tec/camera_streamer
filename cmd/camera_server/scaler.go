package main

//type Scaler struct {
//	gst.Element
//	//*gst.BaseBin
//	targetSize int
//}
//
//var resToCapString = map[int]string{
//	720:  "video/x-raw, width=1280, height=720",
//	1080: "video/x-raw, width=1920, height=1080",
//}
//
//func NewScaler(name string, targetSize int) (Scaler, error) {
//	//scalerBin, err := gst.NewBaseBin(fmt.Sprintf("%s-%i-scaler", Id, targetSize))
//	//if err != nil {
//	//	return Scaler{}, err
//	//}
//
//	//scale, err := gst.NewVideoScale(fmt.Sprintf("%s-%i-scaler-scale", Id, targetSize))
//	//if err != nil {
//	//	return Scaler{}, err
//	//}
//
//	//filter, err := gst.NewCapsFilter(fmt.Sprintf("%s-%i-scaler-filter", Id, targetSize))
//	//if err != nil {
//	//	return Scaler{}, err
//	//}
//
//	encoder, err := gst.NewX264Enc(fmt.Sprintf("%s-%i-scaler-vp8enc", name, targetSize))
//	if err != nil {
//		return Scaler{}, err
//	}
//
//	//capsString, ok := resToCapString[targetSize]
//	//if !ok {
//	//	return Scaler{}, fmt.Errorf("unknown target size")
//	//}
//
//	//filterCaps, err := gst.NewBaseCapsFromString(capsString)
//	//if err != nil {
//	//	return Scaler{}, err
//	//}
//
//	//filter.SetProperty("caps", filterCaps)
//
//	//scalerBin.AddSinkElement(scale)
//	//scalerBin.AddElement(filter)
//	//scalerBin.AddSrcElement(encoder)
//
//	//gst.LinkElements(scale, filter)
//	//gst.LinkElements(filter, encoder)
//
//	return Scaler{Element: encoder.Element, targetSize: targetSize}, nil
//}

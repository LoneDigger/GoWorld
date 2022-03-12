package bundle

import "me.game/src/size"

const (
	AddBcstCode    = iota //加入廣播
	RemoveReqCode         //移除請求
	RemoveBcstCode        //移除廣播
	DegreeReqCode         //轉角請求
	DegreeAckCode         //轉角結果
	MoveBcstCode          //移動廣播
	AddAckCode            //加入回應
	RejectCode            //請求拒絕
	AddReqCode            //加入請求
	EchoReqCode           //心跳請求
	EchoAckCode           //心跳回應
	CloseCode             //關閉連線
)

// 加入請求
type AddRequest struct {
	Name string `json:"name"`
}

// 加入廣播
type AddBroadcast struct {
	ID    uint32     `json:"id"`
	Name  string     `json:"name"`
	Point size.Point `json:"point"`
}

// 解析用
type Broadcast struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
}

// 關閉
type CloseResponse struct {
}

// 移動請求
type DegreeRequest struct {
	Degree float64 `json:"degree"`
}

// 移動回應
type DegreeResponse struct {
	Point size.Point `json:"point"`
}

// 心跳請求
type EchoRequest struct {
}

// 心跳回應
type EchoResponse struct {
}

// 移動廣播
type PosBroadcast struct {
	ID    uint32     `json:"id"`
	Point size.Point `json:"point"`
}

// 拒絕回應
type RejectResponse struct {
	RejectCode int `json:"reject"`
}

// 離開請求
type RemoveRequest struct {
}

// 移除玩家廣播
type RemoveBroadcast struct {
	ID uint32 `json:"id"`
}

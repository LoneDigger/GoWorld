package utils

import (
	"hash/crc32"
	"time"
)

// 讀取限制
const (
	BufferSize   = 1024
	MoveCD       = 10 * time.Millisecond  //移動CD
	MoveDistance = 10                     //移動距離
	AreaDistance = 1                      //可見區域
	NameTimeout  = 300 * time.Millisecond //詢問逾時
	ReadTimeout  = 3 * time.Minute        //讀取逾時
	WriteTimeout = 50 * time.Millisecond  //寫入逾時
	ChannelCount = 8                      //佇列長度
	EchoSecond   = 10                     //回應秒數 10秒
	UpdateCD     = 30 * time.Millisecond  //更新
)

func Crc32(ip string) uint32 {
	return crc32.ChecksumIEEE([]byte(ip))
}

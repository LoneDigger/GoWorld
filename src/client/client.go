package client

import (
	"context"
	"encoding/json"
	"math"
	"net"
	"time"

	"github.com/sasha-s/go-deadlock"
	"github.com/sirupsen/logrus"
	"me.game/src/bundle"
	"me.game/src/session"
	"me.game/src/size"
	"me.game/src/utils"
)

type Client struct {
	id         uint32                   //玩家編號
	name       string                   //顯示名稱
	position   size.Point               //當前座標
	vision     *Vision                  //可見區域
	moveCh     chan bundle.PosBroadcast //移動通道
	exitCh     chan uint32              //離開通道通道
	moveUpdate time.Time                //上次更新時間
	echoUpdate time.Time                //回應時間
	lock       deadlock.Mutex           //鎖
	closed     bool                     //已關閉
	prot       session.SessionHandle    //通訊
	cancel     context.CancelFunc       //取消 go
	updateCh   chan *Client             //狀態更新通道
}

func NewClient(id uint32, name string, moveCh chan bundle.PosBroadcast, exitCh chan uint32,
	updateCh chan *Client, prot session.SessionHandle) *Client {
	return &Client{
		id:         id,
		name:       name,
		prot:       prot,
		vision:     NewVision(),
		moveCh:     moveCh,
		exitCh:     exitCh,
		moveUpdate: time.Now(),
		echoUpdate: time.Now(),
		closed:     false,
		updateCh:   updateCh,
	}
}

// 加入視野
func (c *Client) AddAreas(id int) {
	c.vision.Add(id)
}

// 主區域
func (c *Client) AreaID() int {
	return c.vision.main
}

// 取得視野
func (c *Client) Areas() AreaSet {
	return c.vision.Areas()
}

// 移動
func (c *Client) callMove(angle float64) {
	if time.Since(c.moveUpdate) < utils.MoveCD {
		c.Send(bundle.Broadcast{
			Code: bundle.RejectCode,
			Message: bundle.RejectResponse{
				RejectCode: bundle.DegreeReqCode,
			},
		})
		return
	}

	radian := math.Pi / 180.0 * angle

	var pos size.Point
	pos.X = int(math.Cos(radian)*utils.MoveDistance) + c.position.X
	pos.Y = int(math.Sin(radian)*utils.MoveDistance) + c.position.Y

	c.moveCh <- bundle.PosBroadcast{
		ID:    c.id,
		Point: pos,
	}
}

func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return
	}

	logrus.WithField("client_id", c.id).Info("close")

	c.cancel()
	err := c.prot.Close()

	if err != nil {
		logrus.WithError(err).Error()
	}
	c.closed = true
}

// 回應時間(秒)
func (c *Client) EchoTime() int {
	return int(time.Since(c.echoUpdate).Seconds())
}

// 玩家編號
func (c *Client) ID() uint32 {
	return c.id
}

// 初始化
func (c *Client) Init(areaID int, pos size.Point) {
	c.position = pos
	c.vision.SetMain(areaID)
	c.moveUpdate = time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.receive()
	go c.updateLoop(ctx)

	c.Send(bundle.Broadcast{
		Code: bundle.AddAckCode,
		Message: bundle.PosBroadcast{
			ID:    c.id,
			Point: pos,
		},
	})
}

// 玩家名稱
func (c *Client) Name() string {
	return c.name
}

// 座標
func (c *Client) Position() size.Point {
	return c.position
}

// 接收封包
func (c *Client) receive() {
	var err error
	var object json.RawMessage
	b := bundle.Broadcast{
		Message: &object,
	}

	for {
		//io: read/write on closed pipe
		err = c.prot.Read(&b)
		if err != nil {
			c.cancel()
			return
		}

		switch b.Code {

		case bundle.DegreeReqCode: //移動
			var move bundle.DegreeRequest
			err = json.Unmarshal(object, &move)
			if err == nil {
				c.callMove(move.Degree)
			}

		case bundle.EchoAckCode: //心跳回應
			var echo bundle.EchoResponse
			err = json.Unmarshal(object, &echo)
			if err == nil {
				c.echoUpdate = time.Now()
			}

		case bundle.CloseCode: //關閉
			c.exitCh <- c.id
			return
		}
	}
}

// 移除視野
func (c *Client) RemoveArea(id int) {
	c.vision.Remove(id)
}

func (c *Client) Send(b bundle.Broadcast) {
	c.lock.Lock()

	if c.closed {
		c.lock.Unlock()
		return
	}

	err := c.prot.Write(b)
	c.lock.Unlock()

	if err != nil {
		if err.Error() == "timeout" {
		} else if e, ok := err.(net.Error); ok && e.Timeout() {
		} else {
			// io: read/write on closed pipe
			logrus.WithError(err).WithField("addr", c.prot.RemoteAddr()).Error("sendLoop")
			c.Close()
		}
	}
}

// 送出關閉
func (c *Client) SendClose() {
	c.Send(bundle.Broadcast{
		Code:    bundle.CloseCode,
		Message: bundle.CloseResponse{},
	})
}

// 變更中心
func (c *Client) SetAreaID(id int) {
	c.vision.SetMain(id)
}

// 設定座標
func (c *Client) SetPosition(p size.Point) {
	c.position = p
	c.moveUpdate = time.Now()

	//送出新座標
	c.Send(bundle.Broadcast{
		Code: bundle.DegreeAckCode,
		Message: bundle.DegreeResponse{
			Point: p,
		},
	})
}

func (c *Client) updateLoop(ctx context.Context) {
	t := time.NewTicker(utils.UpdateCD)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.updateCh <- c
			t.Reset(utils.UpdateCD)
		}
	}
}

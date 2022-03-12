package module

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	"me.game/src/bundle"
	"me.game/src/client"
	"me.game/src/session"
	"me.game/src/size"
	"me.game/src/utils"
)

type world struct {
	name     string
	split    int                      //切割區塊數量
	areas    [][]*Area                //區塊
	size     size.Point               //總大小
	areaSize size.Point               //區塊大小
	moveCh   chan bundle.PosBroadcast //移動
	exitCh   chan uint32              //離開
	players  *client.Players
}

func newWorld(name string, p size.Point, split int) *world {
	var areaSize size.Point
	areaSize.X = p.X / split
	areaSize.Y = p.Y / split

	areas := make([][]*Area, split)
	for i := 0; i < split; i++ {
		areas[i] = make([]*Area, split)
		for o := 0; o < split; o++ {
			id := i + o*split
			areas[i][o] = NewAres(id)
		}
	}

	return &world{
		name:     name,
		split:    split,
		size:     p,
		areaSize: areaSize,
		areas:    areas,
		moveCh:   make(chan bundle.PosBroadcast, utils.ChannelCount),
		exitCh:   make(chan uint32, utils.ChannelCount),
		players:  client.NewPlayer(),
	}
}

// 加入世界
func (w *world) Add(id uint32, name string, prot session.SessionHandle) {
	client := client.NewClient(id, name, w.moveCh, w.exitCh, prot)

	//隨機位置
	p := size.Point{
		X: rand.Intn(w.size.X),
		Y: rand.Intn(w.size.Y),
	}

	posX := p.X / w.areaSize.X
	posY := p.Y / w.areaSize.Y

	area := w.areas[posX][posY]
	client.Init(area.ID(), p)

	pos := w.roundArea(posX, posY)
	for _, p := range pos {
		if p.X == posX && p.Y == posY {
			w.areas[p.X][p.Y].AddClient(client)
		} else {
			w.areas[p.X][p.Y].AddBroadcast(client)
		}
		client.AddAreas(w.areaPosToID(p))
	}

	w.players.AddClient(client)

	logrus.WithFields(logrus.Fields{
		"client_id": client.ID(),
		"world":     w.name,
	})
}

// 限制大地圖邊界
func (w *world) bounds(p size.Point) size.Point {
	var pos size.Point

	if p.X < 0 {
		pos.X = 0
	} else if p.X >= w.size.X {
		pos.X = w.size.X - 1
	} else {
		pos.X = p.X
	}

	if p.Y < 0 {
		pos.Y = 0
	} else if p.Y >= w.size.Y {
		pos.Y = w.size.Y - 1
	} else {
		pos.Y = p.Y
	}

	return pos
}

func (w *world) callMove(m bundle.PosBroadcast) {
	client, ok := w.players.Get(m.ID)
	if !ok {
		logrus.WithField("client_id", m.ID).Warn("move_fail")
		return
	}

	m.Point = w.bounds(m.Point)
	newX := m.Point.X / w.areaSize.X
	newY := m.Point.Y / w.areaSize.Y

	newAreaID := w.posToAreaID(m.Point)
	oldAreaID := client.AreaID()

	if oldAreaID != newAreaID {
		oldArray, newArray := w.fog(client.Position(), m.Point, client.Areas())
		for _, p := range oldArray {
			w.areas[p.X][p.Y].RemoveBroadcast(client)
			client.RemoveArea(w.areaPosToID(p))
		}
		//移除舊區域
		oldPos := w.idToPos(client.AreaID())
		w.areas[oldPos.X][oldPos.Y].RemoveClient(client)

		for _, p := range newArray {
			w.areas[p.X][p.Y].AddBroadcast(client)
			client.AddAreas(w.areaPosToID(p))
		}
		//新增新區域
		w.areas[newX][newY].AddClient(client)

		client.SetAreaID(newAreaID)
	}

	//新座標
	client.SetPosition(m.Point)
	w.distribute(newX, newY, m)
}

// 移除玩家
func (w *world) callRemove(client *client.Client) {
	center := w.idToPos(client.AreaID())
	w.areas[center.X][center.Y].RemoveClient(client)

	pos := w.roundArea(center.X, center.Y)
	for _, p := range pos {
		w.areas[p.X][p.Y].RemoveBroadcast(client)
	}

	client.Close()
	w.players.RemoveClient(client.ID())
	logrus.WithField("client_id", client.ID()).Info("exit_world")
}

// 確認回應
func (w *world) checkEcho() {
	array := w.players.CheckEcho()
	for _, c := range array {
		c.SendClose()
		w.callRemove(c)
	}
}

// 轉發移動
func (w *world) distribute(posX, posY int, p bundle.PosBroadcast) {
	arr := w.roundArea(posX, posY)
	for _, a := range arr {
		w.areas[a.X][a.Y].Move(p)
	}
}
func (w *world) exitLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case id := <-w.exitCh:
			client, ok := w.players.Get(id)
			if ok {
				w.callRemove(client)
			} else {
				logrus.WithField("client_id", id).Warn("remove_fail")
			}
		}
	}
}

// 有無新增加區塊
func (w *world) fog(oldPos, newPos size.Point, oldArea client.AreaSet) ([]size.Point, []size.Point) {
	oldPosition := make([]size.Point, 0)
	newPosition := make([]size.Point, 0)

	newArea := make(map[int]size.Point)
	newArray := w.roundArea(newPos.X/w.areaSize.X, newPos.Y/w.areaSize.Y)

	for _, p := range newArray {
		if _, ok := oldArea[w.areaPosToID(p)]; !ok {
			//表示新加入的區域
			newPosition = append(newPosition, p)
		}

		newArea[w.areaPosToID(p)] = p
	}

	for id := range oldArea {
		if _, ok := newArea[id]; !ok {
			//表示舊的區域
			oldPosition = append(oldPosition, w.idToPos(id))
		}
	}

	return oldPosition, newPosition
}

// 心跳機制
func (w *world) heartbeat(ctx context.Context) {
	askT := time.NewTimer(20 * time.Second)
	checkT := time.NewTimer(30 * time.Second)

	defer askT.Stop()
	defer checkT.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-askT.C:
			logrus.Info("send_echo_start")
			w.sendEcho()
			logrus.Info("send_echo_end")

		case <-checkT.C:
			logrus.Info("check_echo_start")
			w.checkEcho()
			logrus.Info("check_echo_end")

			askT.Reset(20 * time.Second)
			checkT.Reset(30 * time.Second)
		}
	}
}

// 編號轉成區域位置
func (w *world) idToPos(id int) size.Point {
	return size.Point{
		X: id % w.split,
		Y: id / w.split,
	}
}

// 移動處裡
func (w *world) moveLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case m := <-w.moveCh:
			w.callMove(m)
		}
	}
}

// 心跳詢問
func (w *world) sendEcho() {
	b := bundle.Broadcast{
		Code:    bundle.EchoReqCode,
		Message: bundle.EchoRequest{},
	}

	w.players.Send(b)
}

func (w *world) Start(ctx context.Context) {
	for i := 0; i < w.split; i++ {
		for o := 0; o < w.split; o++ {
			w.areas[i][o].Start(ctx)
		}
	}

	go w.moveLoop(ctx)
	go w.exitLoop(ctx)
	go w.heartbeat(ctx)
}

// 取得周圍
func (w *world) roundArea(posX, posY int) []size.Point {
	arr := make([]size.Point, 0)

	for x := posX - utils.AreaDistance; x <= posX+utils.AreaDistance; x++ {
		if x < 0 || x >= w.split {
			continue
		}

		for y := posY - utils.AreaDistance; y <= posY+utils.AreaDistance; y++ {
			if y < 0 || y >= w.split {
				continue
			}

			arr = append(arr, size.Point{
				X: x,
				Y: y,
			})
		}
	}
	return arr
}

// 座標轉成編號
func (w *world) posToAreaID(p size.Point) int {
	return w.areaPosToID(size.Point{
		X: p.X / w.areaSize.X,
		Y: p.Y / w.areaSize.Y,
	})
}

// 區域位置轉成編號
func (w *world) areaPosToID(p size.Point) int {
	return p.X + p.Y*w.split
}

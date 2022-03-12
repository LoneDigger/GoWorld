package module

import (
	"context"

	"me.game/src/bundle"
	"me.game/src/client"
	"me.game/src/utils"
)

// 世界中一小區域
type Area struct {
	id      int //區域編號
	players *client.Players

	addCh    chan *client.Client
	removeCh chan *client.Client

	addBcst    chan *client.Client
	removeBcst chan *client.Client

	moveCh chan bundle.PosBroadcast
}

func NewAres(id int) *Area {
	return &Area{
		id:         id,
		players:    client.NewPlayer(),
		addCh:      make(chan *client.Client, utils.ChannelCount),
		removeCh:   make(chan *client.Client, utils.ChannelCount),
		addBcst:    make(chan *client.Client, utils.ChannelCount),
		removeBcst: make(chan *client.Client, utils.ChannelCount),
		moveCh:     make(chan bundle.PosBroadcast, utils.ChannelCount),
	}
}

// 加入廣播
func (a *Area) AddBroadcast(client *client.Client) {
	a.addBcst <- client
}

// 區域進入一個玩家
func (a *Area) AddClient(client *client.Client) {
	a.addCh <- client
}

// 廣播處理
func (a *Area) bcstLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case c1 := <-a.addBcst:
			a.players.AddBroadcast(c1)

		case c1 := <-a.removeBcst:
			a.players.RemoveBroadcast(c1)
		}
	}
}

// 進出區域
func (a *Area) clientLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case c := <-a.addCh:
			a.players.AddClient(c)

		case c := <-a.removeCh:
			a.players.RemoveClient(c.ID())
		}
	}
}

// 取得編號
func (a *Area) ID() int {
	return a.id
}

// 移動廣播
func (a *Area) Move(pos bundle.PosBroadcast) {
	a.moveCh <- pos
}

func (a *Area) moveLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case pos := <-a.moveCh: //玩家移動廣播
			b := bundle.Broadcast{
				Code:    bundle.MoveBcstCode,
				Message: pos,
			}

			a.players.Move(pos.ID, b)
		}
	}
}

// 移除廣播
func (a *Area) RemoveBroadcast(client *client.Client) {
	a.removeBcst <- client
}

// 移除玩家
func (a *Area) RemoveClient(client *client.Client) {
	a.removeCh <- client
}

// 啟動
func (a *Area) Start(ctx context.Context) {
	go a.bcstLoop(ctx)
	go a.clientLoop(ctx)
	go a.moveLoop(ctx)
}

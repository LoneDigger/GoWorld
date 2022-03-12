package module

import (
	"context"

	"me.game/src/client"
	"me.game/src/utils"
)

// 區域
type Area struct {
	id         int                 //區域編號
	players    *client.Players     //當前區域內玩家
	addCh      chan *client.Client //進入通道
	removeCh   chan *client.Client //出來通道
	addBcst    chan *client.Client //進入廣播通道
	removeBcst chan *client.Client //出來廣播通道
}

func NewAres(id int) *Area {
	return &Area{
		id:         id,
		players:    client.NewPlayer(),
		addCh:      make(chan *client.Client, utils.ChannelCount),
		removeCh:   make(chan *client.Client, utils.ChannelCount),
		addBcst:    make(chan *client.Client, utils.ChannelCount),
		removeBcst: make(chan *client.Client, utils.ChannelCount),
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

// 複製玩家清單
func (a *Area) Clone() []*client.Client {
	return a.players.Clone()
}

// 移除廣播
func (a *Area) RemoveBroadcast(client *client.Client) {
	a.removeBcst <- client
}

// 移除玩家
func (a *Area) RemoveClient(client *client.Client) {
	a.removeCh <- client
}

func (a *Area) Start(ctx context.Context) {
	go a.bcstLoop(ctx)
	go a.clientLoop(ctx)
}

func (a *Area) Stop() {
	close(a.addCh)
	close(a.removeCh)
	close(a.addBcst)
	close(a.removeBcst)
}

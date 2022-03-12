package module

import (
	"context"
	"math/rand"
	"strconv"

	"me.game/src/config"
	"me.game/src/session"
)

type CoreHandle interface {
	Start()
	Add(uint32, string, session.SessionHandle)
	Stop()
}

type core struct {
	worlds []*world
	cancel context.CancelFunc
}

func NewCore(cfg config.Config) CoreHandle {
	worlds := make([]*world, cfg.World)
	for i := 0; i < cfg.World; i++ {
		worlds[i] = newWorld(strconv.Itoa(i), cfg.Size, cfg.Split)
	}

	return &core{
		worlds: worlds,
	}
}

func (c *core) Add(id uint32, name string, prot session.SessionHandle) {
	rnd := rand.Intn(len(c.worlds))
	c.worlds[rnd].Add(id, name, prot)
}

func (c *core) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	for _, w := range c.worlds {
		w.Start(ctx)
	}
}

func (c *core) Stop() {
	c.cancel()
}

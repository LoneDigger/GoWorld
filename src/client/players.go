package client

import (
	"github.com/sasha-s/go-deadlock"
	"github.com/sirupsen/logrus"
	"me.game/src/bundle"
	"me.game/src/utils"
)

type Players struct {
	m    map[uint32]*Client
	lock deadlock.Mutex
}

func NewPlayer() *Players {
	return &Players{
		m: make(map[uint32]*Client),
	}
}

func (p *Players) AddClient(client *Client) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.m[client.id] = client
}

func (p *Players) AddBroadcast(c1 *Client) {
	p.lock.Lock()
	defer p.lock.Unlock()

	id1 := c1.ID()
	b := bundle.Broadcast{
		Code: bundle.AddBcstCode,
		Message: bundle.AddBroadcast{
			ID:    id1,
			Name:  c1.Name(),
			Point: c1.Position(),
		},
	}

	for _, c2 := range p.m {
		id2 := c2.ID()
		if id1 != id2 {
			c2.Send(b)

			c1.Send(bundle.Broadcast{
				Code: bundle.AddBcstCode,
				Message: bundle.AddBroadcast{
					ID:    id2,
					Name:  c2.Name(),
					Point: c2.Position(),
				},
			})
		}
	}
}

func (p *Players) CheckEcho() []*Client {
	p.lock.Lock()
	defer p.lock.Unlock()

	array := make([]*Client, 0)
	for _, c := range p.m {
		if c.EchoTime() > utils.EchoSecond {
			logrus.WithField("client_id", c.ID()).Warn("no_echo")
			array = append(array, c)
		}
	}
	return array
}

func (p *Players) Clone() []*Client {
	p.lock.Lock()
	defer p.lock.Unlock()

	array := make([]*Client, len(p.m))
	i := 0
	for _, c := range p.m {
		array[i] = c
		i++
	}
	return array
}

func (p *Players) Get(id uint32) (*Client, bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	value, ok := p.m[id]
	return value, ok
}

func (p *Players) Move(id uint32, b bundle.Broadcast) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, c := range p.m {
		if id != c.id {
			c.Send(b)
		}
	}
}

func (p *Players) RemoveBroadcast(c1 *Client) {
	p.lock.Lock()
	defer p.lock.Unlock()

	id1 := c1.ID()

	b := bundle.Broadcast{
		Code: bundle.RemoveBcstCode,
		Message: bundle.RemoveBroadcast{
			ID: id1,
		},
	}

	for _, c2 := range p.m {
		id2 := c2.ID()
		if id1 != id2 {
			c2.Send(b)

			c1.Send(bundle.Broadcast{
				Code: bundle.RemoveBcstCode,
				Message: bundle.RemoveBroadcast{
					ID: c2.ID(),
				},
			})
		}
	}
}

func (p *Players) RemoveClient(id uint32) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.m, id)
}

func (p *Players) Send(b bundle.Broadcast) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, c := range p.m {
		c.Send(b)
	}
}

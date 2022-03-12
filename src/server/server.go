package server

import (
	"errors"
	"strings"

	"me.game/src/config"
)

type serverHandle interface {
	Start() error
	Stop() error
}

func NewProtocol(c config.Config) (serverHandle, error) {
	switch strings.ToLower(c.Protocol) {
	case "kcp":
		return newKcp(c)

	case "websocket":
		return newHttp(c)

	default:
		return nil, errors.New("no protocal")
	}
}

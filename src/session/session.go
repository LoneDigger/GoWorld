package session

import (
	"errors"

	"me.game/src/bundle"
)

var errorName = errors.New("named ?")

type SessionHandle interface {
	Close() error
	Read(*bundle.Broadcast) error
	ReadName() (string, error)
	RemoteAddr() string
	Write(bundle.Broadcast) error
}

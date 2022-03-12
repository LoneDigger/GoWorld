package session

import (
	"encoding/json"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"me.game/src/bundle"
	"me.game/src/utils"
)

type kcpProt struct {
	session *kcp.UDPSession
	buffer  []byte
}

func NewKcp(s *kcp.UDPSession) (SessionHandle, error) {
	s.SetWindowSize(4096, 4096)
	s.SetWriteDelay(false)
	s.SetACKNoDelay(false)
	s.SetNoDelay(1, 100, 2, 0)

	return &kcpProt{
		session: s,
		buffer:  make([]byte, 1024),
	}, nil
}

func (s *kcpProt) Close() error {
	return s.session.Close()
}

func (s *kcpProt) Read(b *bundle.Broadcast) error {
	return s.readJSON(b)
}

func (s *kcpProt) ReadName() (string, error) {
	var object json.RawMessage
	b := bundle.Broadcast{
		Message: &object,
	}

	s.session.SetReadDeadline(time.Now().Add(utils.NameTimeout))
	err := s.readJSON(&b)
	if err != nil {
		return "", err
	}

	if b.Code == bundle.AddReqCode {
		var add bundle.AddRequest
		err = json.Unmarshal(object, &add)
		if err != nil {
			return "", err
		}

		s.session.SetReadDeadline(time.Now().Add(utils.ReadTimeout))
		return add.Name, nil
	}

	return "", errorName
}

func (s *kcpProt) readJSON(b *bundle.Broadcast) error {
	n, err := s.session.Read(s.buffer)
	if err != nil {
		return err
	}
	return json.Unmarshal(s.buffer[:n], b)
}

func (s *kcpProt) RemoteAddr() string {
	return s.session.RemoteAddr().String()
}

func (s *kcpProt) Write(b bundle.Broadcast) error {
	return s.writeJSON(b)
}

func (s *kcpProt) writeJSON(b bundle.Broadcast) error {
	buffer, err := json.Marshal(b)
	if err != nil {
		return err
	}

	s.session.SetWriteDeadline(time.Now().Add(utils.WriteTimeout))
	_, err = s.session.Write(buffer)
	return err
}

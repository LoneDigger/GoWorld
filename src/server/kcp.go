package server

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go/v5"
	"me.game/src/config"
	"me.game/src/module"
	"me.game/src/session"
	"me.game/src/utils"
)

// var password = []byte("Arw8ptUoEe")
// var salt = []byte("lsXSCvYDnn")

type kcpServer struct {
	listener *kcp.Listener
	core     module.CoreHandle
	cancel   context.CancelFunc
	acceptCh chan *kcp.UDPSession
}

func newKcp(cfg config.Config) (serverHandle, error) {
	return &kcpServer{
		core:     module.NewCore(cfg),
		acceptCh: make(chan *kcp.UDPSession, utils.ChannelCount),
	}, nil
}

// 加
func (s *kcpServer) addWorld(l *kcp.UDPSession) {
	prot, err := session.NewKcp(l)
	if err != nil {
		logrus.WithError(err).Error("addWorld")
		return
	}

	name, err := prot.ReadName()
	if err != nil {
		if err.Error() != "timeout" {
			logrus.WithError(err).WithField("addr", prot.RemoteAddr()).Error("ReadName")
		}
		prot.Close()
		return
	}

	id := utils.Crc32(l.RemoteAddr().String())
	s.core.Add(id, name, prot)
	logrus.WithField("client_id", id).Info("add_world")
}

func (s *kcpServer) Start() error {
	// key := pbkdf2.Key(password, salt, 1024, 32, sha1.New)
	// block, _ := kcp.NewAESBlockCrypt(key)
	// listener, err := kcp.ListenWithOptions(":6060", block, 0, 0)

	listener, err := kcp.ListenWithOptions(":6060", nil, 0, 0)
	if err != nil {
		return err
	}

	s.listener = listener
	s.core.Start()

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.acceptLoop(ctx)

	logrus.WithField("protocol", "kcp").Info("start")
	for {
		l, err := listener.AcceptKCP()
		if err != nil {
			return err
		}
		s.acceptCh <- l
	}
}

func (s *kcpServer) Stop() error {
	s.cancel()
	s.core.Stop()
	return s.listener.Close()
}

// 玩家請求連線
func (s *kcpServer) acceptLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case l := <-s.acceptCh:
			go s.addWorld(l)
		}
	}
}

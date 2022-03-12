package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"me.game/src/config"
	"me.game/src/module"
	"me.game/src/session"
	"me.game/src/utils"
)

type httServer struct {
	eng  *gin.Engine
	srv  *http.Server
	core module.CoreHandle
}

func newHttp(cfg config.Config) (serverHandle, error) {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: engine,
	}

	return &httServer{
		eng:  engine,
		srv:  srv,
		core: module.NewCore(cfg),
	}, nil
}

// 玩家加入世界
func (s *httServer) addWorld(c *gin.Context) {
	prot, err := session.NewWebSocket(c.Writer, c.Request)
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

	// create player
	id := utils.Crc32(c.Request.RemoteAddr)
	s.core.Add(id, name, prot)
}

func (s *httServer) Start() error {
	//ping
	s.eng.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	s.eng.GET("/add", s.addWorld)

	s.core.Start()

	logrus.WithField("protocol", "websocket").Info("start")
	return s.srv.ListenAndServe()
}

func (s *httServer) Stop() error {
	s.core.Stop()
	return s.srv.Shutdown(context.Background())
}

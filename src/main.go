package main

import (
	"math/rand"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"me.game/src/config"
	"me.game/src/server"
)

func main() {

	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	logrus.SetLevel(logrus.DebugLevel)

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())

	var cfg config.Config
	cfg.Protocol = "kcp"
	cfg.Port = 6060
	cfg.Size.X = 1000
	cfg.Size.Y = 1000
	cfg.Split = 10
	cfg.World = 1

	s, err := server.NewProtocol(cfg)
	if err != nil {
		logrus.WithError(err).Error()
		return
	}

	s.Start()
}

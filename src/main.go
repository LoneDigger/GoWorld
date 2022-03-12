package main

import (
	"math/rand"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sasha-s/go-deadlock"
	"github.com/sirupsen/logrus"
	"me.game/src/config"
	"me.game/src/server"
)

// https://iter01.com/594099.html
// https://github.com/gorilla/websocket/tree/master/examples/chat
func main() {
	// runtime.SetMutexProfileFraction(1)
	// runtime.SetBlockProfileRate(1)
	//go http.ListenAndServe("localhost:80", nil)

	// f, _ := os.OpenFile("cpu.profile", os.O_CREATE|os.O_RDWR, 0644)
	// pprof.StartCPUProfile(f)

	gin.ForceConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	logrus.SetLevel(logrus.DebugLevel)

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().Unix())

	deadlock.Opts.DeadlockTimeout = time.Second * 10
	//deadlock.Opts.Disable = true

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

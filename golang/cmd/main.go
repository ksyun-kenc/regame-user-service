package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	"regame-user-service/config"
	"regame-user-service/server"
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()
	defer glog.Flush()

	var err error
	cfg := &config.DefaultConfig
	if len(os.Args) > 1 {
		cfg, err = config.ParseConfigFromJsonFile(os.Args[1])
		if err != nil {
			glog.Errorf("ParseConfigFromJsonFile err %s\n", err)
			return
		}
	}
	err = cfg.Validate()
	if err != nil {
		glog.Errorf("config Validate err %s\n", err)
		return
	}
	cfg.Print()

	s, err := server.New(cfg)
	if err != nil {
		glog.Errorf("server.New error %s\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		select {
		case <-ctx.Done():
			return
		case <-interrupt:
			s.Stop()
		}
	}()

	err = s.Run(ctx)
	if err != nil {
		glog.Errorf("server.Run exit with error %s\n", err)
	}
}

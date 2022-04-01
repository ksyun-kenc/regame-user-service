package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"

	"regame-user-service/config"
	"regame-user-service/service"
)

func init() {
	_ = flag.Set("logtostderr", "true")
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Usage:", os.Args[0], "<config-file-path>")
		return
	}

	flag.Parse()
	defer glog.Flush()

	cfg, err := config.ParseConfigFile(os.Args[1])
	if err != nil {
		glog.Errorf("ParseConfigFile failed: %s\n", err)
		return
	}
	err = cfg.Validate()
	if err != nil {
		glog.Errorf("Invalid config: %s\n", err)
		return
	}
	cfg.Print()

	svc, err := service.New(cfg)
	if err != nil {
		glog.Errorf("service.New failed: %s\n", err)
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
			svc.Stop()
		}
	}()

	err = svc.Run(ctx)
	if err != nil {
		glog.Errorf("service.Run exit with error: %s\n", err)
	}
}

package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"regame-user-service/config"
	"regame-user-service/database"
)

const (
	cleanSessionInterval = 10 * time.Second // second
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc

	cfg *config.Config

	mux        *http.ServeMux
	httpServer *http.Server

	userService UserService
}

func (s *Server) CleanExpiredUserSession() {
	t := time.NewTicker(cleanSessionInterval)
	defer t.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
		}

		s.userService.CleanExpiredSession()
	}
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
		userService: UserService{
			cfg:      &cfg.UserService,
			sessions: make(map[string]*UserSession),
		},
	}
	return s, nil
}

func (s *Server) registerHandler() {
	s.mux.HandleFunc("/user", s.userService.UserHandler)
}

func (s *Server) Run(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	database.TiKVClientInit(s.cfg.UserService.PdAddress)

	s.mux = http.NewServeMux()
	s.httpServer = &http.Server{Handler: s.mux}
	s.registerHandler()

	go s.CleanExpiredUserSession()

	go func() {
		select {
		case <-s.ctx.Done():
			_ = s.httpServer.Close()
		}
	}()
	return s.httpServer.Serve(l)
}

func (s *Server) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

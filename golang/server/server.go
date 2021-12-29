package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"regame-user-service/config"
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

	sync.Mutex
	sessions map[string]*Session
}

func isTimeExpired(t time.Time, expiredDuration time.Duration) bool {
	n := time.Now()
	if n.After(t.Add(expiredDuration)) {
		return true
	}
	return false
}

func (s *Server) CleanExpiredSession() {
	t := time.NewTicker(cleanSessionInterval)
	defer t.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-t.C:
		}

		s.Lock()
		for id, session := range s.sessions {
			if isTimeExpired(session.UpdateTime, s.cfg.ExpiredDuration) {
				delete(s.sessions, id)
			}
		}
		s.Unlock()
	}
}

func New(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
	return s, nil
}

func (s *Server) registerHandler() {
	s.mux.HandleFunc("/user", s.UserHandler)
}

func (s *Server) Run(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.mux = http.NewServeMux()
	s.httpServer = &http.Server{Handler: s.mux}
	s.registerHandler()

	go s.CleanExpiredSession()

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

package main

import (
	"net"
	"time"
)

type Server struct {
	MirrorConfig MirrorConfig
	Listener     net.Listener
}

func NewServer(mc MirrorConfig) (*Server, error) {
	var (
		err      error
		network  string
		address  string
		listener net.Listener
	)

	network, address = ParseConfigAddress(mc.Local)

	if listener, err = net.Listen(network, address); err != nil {
		return nil, err
	}

	return &Server{Listener: listener, MirrorConfig: mc}, nil
}

func (s *Server) AcceptRoutine() {
	for {
		var (
			conn net.Conn
			err  error
		)

		if conn, err = s.Listener.Accept(); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var (
			Sess Session = Session{DownConn: conn, MirrorConfig: s.MirrorConfig}
		)

		Sess.Start()
	}
}

func (s *Server) Start() {
	go s.AcceptRoutine()
}

package modbus

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

type (
	Conn struct {
		rwc    net.Conn
		server *Server
		mu     sync.Mutex
	}

	Server struct {
		address string
		serve   func(conn *Conn)
	}
)

func (c *Conn) Read(size int, timeout time.Duration) (*Frame, error) {
	c.rwc.SetReadDeadline(time.Now().Add(timeout))

	defer c.rwc.SetReadDeadline(time.Time{})

	buf := make([]byte, size)

	l, err := c.rwc.Read(buf)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("read", fmt.Sprintf("% X", buf[:l])).Msg("")

	return NewFrame(buf[:l])
}

func (c *Conn) Write(frame *Frame, timeout time.Duration) error {

	c.rwc.SetWriteDeadline(time.Now().Add(timeout))

	defer c.rwc.SetWriteDeadline(time.Time{})

	_, err := c.rwc.Write(frame.Bytes())

	log.Debug().Str("write", fmt.Sprintf("% X", frame.Bytes())).Msg("")

	return err
}

func (c *Conn) Close() error {
	return c.rwc.Close()
}

func (c *Conn) Addr() net.Addr {
	return c.rwc.RemoteAddr()
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
	}
}

func (s *Server) SetServe(serve func(conn *Conn)) {
	s.serve = serve
}

func (s *Server) ListenAndServe() error {
	if s.serve == nil {
		return errors.New("server error: use SetServe of server first")
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	defer listener.Close()
	for {
		rwc, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			s.serve(&Conn{rwc: rwc, server: s})
			_ = rwc.Close()
		}()
	}
}

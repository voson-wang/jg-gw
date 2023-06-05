package modbus

import (
	"errors"
	"net"
	"sync"
	"time"
)

type (
	ErrorLevel uint8

	Conn struct {
		rwc    net.Conn
		server *Server
		mu     sync.Mutex
	}

	Server struct {
		address  string
		serve    func(conn *Conn)
		logLevel ErrorLevel
	}
)

const (
	Silent ErrorLevel = iota
	INFO
	ERROR
	DEBUG
)

func (c *Conn) Read(size int, timeout time.Duration) (*Frame, error) {
	c.rwc.SetReadDeadline(time.Now().Add(timeout))

	defer c.rwc.SetReadDeadline(time.Time{})

	buf := make([]byte, size)

	l, err := c.rwc.Read(buf)
	if err != nil {
		return nil, err
	}

	return NewFrame(buf[:l])
}

func (c *Conn) Write(frame *Frame, timeout time.Duration) error {

	c.rwc.SetWriteDeadline(time.Now().Add(timeout))

	defer c.rwc.SetWriteDeadline(time.Time{})

	_, err := c.rwc.Write(frame.Bytes())

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
		address:  address,
		logLevel: ERROR,
	}
}

func (s *Server) SetServe(serve func(conn *Conn)) {
	s.serve = serve
}

func (s *Server) SetLogLevel(logLevel ErrorLevel) {
	s.logLevel = logLevel
}

func (s *Server) ListenAndServe() error {
	if s.serve == nil {
		return errors.New("use SetServe of server first")
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

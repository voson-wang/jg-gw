package modbus

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestControlRegister(t *testing.T) {
	TestingT(t)
}

type ControlRegisterTestSuite struct{}

var _ = Suite(&ControlRegisterTestSuite{})

func (s *ControlRegisterTestSuite) TestWriteFrame(c *C) {
	writeFrame := Switch.NewWriteFrame(id, []byte{0x00})
	c.Assert(writeFrame.Data, DeepEquals, []byte{0x81, 0x06, 0x00, 0x00, 0x00, 0x01, 0x60, 0x00})
}

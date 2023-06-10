package modbus

import (
	. "gopkg.in/check.v1"
	"testing"
)

func TestActionRegister(t *testing.T) {
	TestingT(t)
}

type ActionRegisterTestSuite struct{}

var _ = Suite(&ActionRegisterTestSuite{})

var id ID = [6]byte{0x07, 0x21, 0x07, 0x63, 0x02, 0x89}

func (s *ActionRegisterTestSuite) TestWriteFrame(c *C) {
	writeFrame := UnderVoltageTripSetting.NewWriteFrame(id, []byte{0xA5, 0x00})
	c.Assert(writeFrame.Data, DeepEquals, []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x42, 0x82, 0x2D, 0x02, 0xA5, 0x00})
}

func (s *ActionRegisterTestSuite) TestReadFrame(c *C) {
	writeFrame := UnderVoltageTripSetting.ReadFrame(id)
	c.Assert(writeFrame.Data, DeepEquals, []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42, 0x82})
}

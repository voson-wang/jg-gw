package main

import (
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"reflect"
	"ricn-smart/jg-gw/modbus"
	"ricn-smart/jg-gw/mq"
	"strings"
	"sync"
)

type getHostRequest struct {
	RequestId string
}

type storage struct {
	m sync.Map
}

var snConn storage

func (s *storage) Load(sn string) (*modbus.Conn, bool) {
	value, ok := s.m.Load(sn)
	if ok {
		return value.(*modbus.Conn), true
	}
	return nil, false
}

func (s *storage) Store(sn string, conn *modbus.Conn) {
	s.m.Store(sn, conn)
}

func (s *storage) Delete(sn string) {
	s.m.Delete(sn)
}

func (s *storage) Range(f func(sn string, conn *modbus.Conn) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(string), value.(*modbus.Conn))
	})
}

// handleMQConn
// mqtt连接上后开始执行订阅
func handleMQConn(client mqtt.Client) {

	// 设备host查询
	if token := client.Subscribe("+/host", mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[0]

		go getHost(sn, client, message.Payload())

	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}

	// 设置属性
	// 相同型号的应用共享请求，使用序列号来找出所在应用
	if token := client.Subscribe(ProjectName+"/+/property/get", mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[1]

		go getProperty(sn, client, message.Payload())

	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}

	// 设置属性
	// 相同型号的应用共享请求，使用序列号来找出所在应用
	if token := client.Subscribe(ProjectName+"/+/property/set", mq.AtMostOnce, func(client mqtt.Client, message mqtt.Message) {
		topic := message.Topic()

		arr := strings.Split(topic, "/")

		sn := arr[1]

		go setProperty(sn, client, message.Payload())

	}); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}
}

type (
	CommonResponse struct {
		RequestId string      `json:"request_id"`
		Success   bool        `json:"success"` // 调用结果是否成功
		Message   string      `json:"message"` // 消息;
		Data      interface{} `json:"data"`    // 实际数据
	}

	setPropertyRequest struct {
		RequestId     string         `json:"request_id"`
		Identifiers   []string       `json:"identifiers"`
		Params        map[string]any `json:"params"`
		ChildDeviceNo string         `json:"child_device_no"`
	}

	getPropertyRequest struct {
		RequestId     string   `json:"request_id"`
		Identifiers   []string `json:"identifiers"`
		ChildDeviceNo string   `json:"child_device_no"`
	}
)

func getHost(sn string, client mqtt.Client, payload []byte) {
	_, ok := snConn.Load(sn)
	if ok {
		var request getHostRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		log.Info().Str("sn", sn).Interface("request", request).Msg("getHost")

		buf, _ := json.Marshal(&CommonResponse{
			RequestId: request.RequestId,
			Success:   true,
			Message:   "OK",
			Data:      GitCommitID,
		})

		if token := client.Publish(request.RequestId, mq.AtMostOnce, false, buf); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Msg("")
		}
	}
}

func (g *getPropertyRequest) Frame() (*modbus.Frame, func(frame *modbus.Frame) (map[string]any, error), error) {
	if len(g.Identifiers) == 0 {
		return nil, nil, errors.New("标识符不能为空")
	}

	id, err := modbus.NewID(g.ChildDeviceNo)
	if err != nil {
		return nil, nil, err
	}

	// 默认只支持单个寄存器
	identifier := g.Identifiers[0]
	register := modbus.FindRegister(identifier)
	if register == nil {
		return nil, nil, fmt.Errorf("找不到匹配的寄存器：%v", identifier)
	}

	ar, ok := register.(*modbus.ActionRegister)
	if ok {
		framer := ar.ReadFrame(id)
		return framer, ar.ParserReadResp, nil
	} else {
		return nil, nil, fmt.Errorf("不支持的寄存器类型，期望：*modbus.ActionRegister，实际：%v", reflect.TypeOf(register))
	}
}

func getProperty(sn string, client mqtt.Client, payload []byte) {
	// 判断sn是否在连接过当前app
	conn, ok := snConn.Load(sn)
	if !ok {
		// 设备未在当前应用上线，忽略请求
		return
	}

	var request getPropertyRequest

	if err := json.Unmarshal(payload, &request); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	log.Info().Str("sn", sn).Interface("request", request).Msg("getProperty")

	frame, parser, err := request.Frame()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	conn.Lock()
	defer conn.Unlock()

	if err := conn.Write(frame, timeout); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	respFrame, err := conn.Read(size, timeout)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	data, err := parser(respFrame)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	buf, _ := json.Marshal(&CommonResponse{
		RequestId: request.RequestId,
		Success:   true,
		Message:   "OK",
		Data:      data,
	})

	log.Info().Str("sn", sn).Interface("request", request).Interface("data", data).Msg("getProperty")

	if token := client.Publish(request.RequestId, mq.AtMostOnce, false, buf); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}
}

func (s *setPropertyRequest) Frame() (*modbus.Frame, func(frame *modbus.Frame) (bool, error), error) {
	if len(s.Identifiers) == 0 {
		return nil, nil, errors.New("标识符不能为空")
	}

	id, err := modbus.NewID(s.ChildDeviceNo)
	if err != nil {
		return nil, nil, err
	}

	// 默认只支持单个寄存器写入
	identifier := s.Identifiers[0]
	register := modbus.FindRegister(identifier)
	if register == nil {
		return nil, nil, fmt.Errorf("找不到匹配的寄存器：%v", identifier)
	}

	var f *modbus.Frame
	var parser func(frame *modbus.Frame) (bool, error)

	switch register.(type) {
	case *modbus.ActionRegister:
		ar := register.(*modbus.ActionRegister)
		val, err := ar.Encode(s.Params)
		if err != nil {
			return nil, nil, err
		}
		f = ar.NewWriteFrame(id, val)
		parser = ar.ParserWriteResp
	case *modbus.ControlRegister:
		cr := register.(*modbus.ControlRegister)
		val, err := cr.Encode(s.Params)
		if err != nil {
			return nil, nil, err
		}
		f = cr.NewWriteFrame(id, val)
		parser = cr.ParserWriteResp
	case modbus.RoRegister:
		return nil, nil, fmt.Errorf("只读寄存器无法写入:%v", identifier)
	default:
		return nil, nil, fmt.Errorf("不支持的寄存器类型:%v", reflect.TypeOf(register))
	}

	return f, parser, nil
}

func setProperty(sn string, client mqtt.Client, payload []byte) {

	// 判断sn是否在连接过当前app
	conn, ok := snConn.Load(sn)
	if !ok {
		// 设备未在当前应用上线，忽略请求
		return
	}

	var request setPropertyRequest

	if err := json.Unmarshal(payload, &request); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	log.Info().Str("sn", sn).Interface("request", request).Msg("setProperty")

	frame, parser, err := request.Frame()
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	conn.Lock()
	defer conn.Unlock()

	if err := conn.Write(frame, timeout); err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	respFrame, err := conn.Read(size, timeout)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	success, err := parser(respFrame)
	if err != nil {
		log.Error().Err(err).Msg("")
		return
	}

	var message string

	if success {
		message = "遥控成功"
	} else {
		message = "遥控失败"
	}

	log.Info().Str("sn", sn).Interface("request", request).Bool("success", success).Msg("setProperty")

	buf, _ := json.Marshal(&CommonResponse{
		RequestId: request.RequestId,
		Success:   success,
		Message:   message,
	})

	if token := client.Publish(request.RequestId, mq.AtMostOnce, false, buf); token.Wait() && token.Error() != nil {
		log.Error().Err(token.Error()).Msg("")
	}
}

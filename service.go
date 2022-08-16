package main

import (
	"e.coding.net/ricnsmart/service/jg-modbus"
	"fmt"
	"jg-gateway/util"
	"reflect"
)

type Service struct {
	Identifier string

	// 由于json解析的缘故，解析后的参数数字类型都是float64，即使原先是uint16、uint32、float32都会被认为是float64
	Encode func(params map[string]interface{}) ([]byte, error)
}

var (
	services = []*Service{bindSwitch}

	bindSwitch = &Service{
		Identifier: "BindSwitch",
		Encode: func(params map[string]interface{}) ([]byte, error) {
			sn, ok := params["SN"].(string)
			if !ok {
				return nil, fmt.Errorf("参数SN类型错误，期望：string，实际：%v", reflect.TypeOf(params["SN"]))
			}

			// 绑定开关约定key为LineNo
			value := params["LineNo"]
			data := make([]byte, 0)
			var l int

			switch value.(type) {
			case []interface{}:
				l = len(value.([]interface{}))
				for _, str := range value.([]interface{}) {
					s, ok := str.(string)
					if !ok {
						return nil, fmt.Errorf("参数LineNo类型错误，期望：[]string，实际：[]%v", reflect.TypeOf(str))
					}
					lineNo, err := util.SetByteSN(s)
					if err != nil {
						return nil, fmt.Errorf("参数LineNo错误，error：[]%v", err)
					}
					data = append(data, lineNo...)
				}
			default:
				return nil, fmt.Errorf("类型 %v 暂不支持", reflect.TypeOf(value))
			}

			address, err := util.SetByteSN(sn)
			if err != nil {
				return nil, fmt.Errorf("参数SN错误，error：[]%v", err)
			}

			cfg := make([]byte, 0)
			cfg = append(cfg, byte(l))

			c := &jg.Frame{
				Function: 0x8f,
				Address:  address,
				Cfg:      cfg,
				Data:     data,
			}
			return c.Bytes(), nil

		},
	}
)

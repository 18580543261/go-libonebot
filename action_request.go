package libonebot

import (
	"errors"
	"strings"

	"github.com/botuniverse/go-libonebot/utils"
	"github.com/tidwall/gjson"
	"github.com/vmihailenco/msgpack/v5"
)

// Request 表示一个动作请求.
type Request struct {
	Action Action      // 动作名称
	Params EasierMap   // 动作参数
	Echo   interface{} // 动作请求的 echo 字段
}

func validateActionRequestMap(m EasierMap) error {
	if action, err := m.GetString("action"); err != nil {
		return errors.New("动作请求 `action` 字段不存在或类型错误")
	} else if action == "" {
		return errors.New("动作请求的 `action` 字段为空")
	}
	if _, err := m.GetMap("params"); err != nil {
		return errors.New("动作请求 `params` 字段不存在或类型错误")
	}
	return nil
}

func parseActionRequestFromMap(prefix string, m map[string]interface{}) (Request, error) {
	em := EasierMapFromMap(m)
	err := validateActionRequestMap(em)
	if err != nil {
		return Request{}, err
	}

	var action Action
	fullname, _ := em.GetString("action")
	prefix_ul := prefix + "_"
	if strings.HasPrefix(fullname, prefix_ul) {
		// extended action
		action = Action{
			Prefix:     prefix,
			Name:       strings.TrimPrefix(fullname, prefix_ul),
			IsExtended: true,
		}
	} else {
		// core action
		action = Action{
			Prefix:     "",
			Name:       fullname,
			IsExtended: false,
		}
	}

	params, _ := em.GetMap("params")
	echo, _ := em.Get("echo")
	r := Request{
		Action: action,
		Params: params,
		Echo:   echo,
	}
	return r, nil
}

func parseActionRequest(prefix string, actionBytes []byte, isBinary bool) (Request, error) {
	var actionRequestMap map[string]interface{}
	if isBinary {
		err := msgpack.Unmarshal(actionBytes, &actionRequestMap)
		if err != nil || actionRequestMap == nil {
			return Request{}, errors.New("动作请求不是一个 MsgPack 映射")
		}
	} else {
		if !gjson.ValidBytes(actionBytes) {
			return Request{}, errors.New("动作请求体不是合法的 JSON")
		}
		m, ok := gjson.Parse(utils.BytesToString(actionBytes)).Value().(map[string]interface{})
		if !ok || m == nil {
			return Request{}, errors.New("动作请求不是一个 JSON 对象")
		}
		actionRequestMap = m
	}
	return parseActionRequestFromMap(prefix, actionRequestMap)
}

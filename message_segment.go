package libonebot

import (
	"encoding/json"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type Segment struct {
	Type string
	Data easierMap
}

func (s Segment) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": s.Type,
		"data": s.Data.Value(),
	})
}

func (s Segment) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(map[string]interface{}{
		"type": s.Type,
		"data": s.Data.Value(),
	})
}

const (
	SegTypeText    = "text"
	SegTypeMention = "mention"
)

func segmentFromMap(m map[string]interface{}) (Segment, error) {
	em := easierMapFromMap(m)
	t, _ := em.GetString("type")
	if t == "" {
		return Segment{}, fmt.Errorf("消息段 `type` 字段不存在或为空")
	}
	data, err := em.GetMap("data")
	if err != nil {
		data = easierMapFromMap(map[string]interface{}{})
	}
	return Segment{
		Type: t,
		Data: data,
	}, nil
}

func (s *Segment) tryMerge(next Segment) bool {
	switch s.Type {
	case SegTypeText:
		if next.Type == SegTypeText {
			text1, err1 := s.Data.GetString("text")
			text2, err2 := next.Data.GetString("text")
			if err1 != nil && err2 == nil {
				s.Data.Set("text", text2)
			} else if err1 == nil && err2 != nil {
				s.Data.Set("text", text1)
			} else if err1 == nil && err2 == nil {
				s.Data.Set("text", text1+text2)
			} else {
				s.Data.Set("text", "")
			}
			return true
		}
	}
	return false
}

func CustomSegment(type_ string, data map[string]interface{}) Segment {
	return Segment{
		Type: type_,
		Data: easierMapFromMap(data),
	}
}

func TextSegment(text string) Segment {
	return CustomSegment(SegTypeText, map[string]interface{}{
		"text": text,
	})
}

func MentionSegment(userID string) Segment {
	return CustomSegment(SegTypeMention, map[string]interface{}{
		"user_id": userID,
	})
}

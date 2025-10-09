package trace

import (
	"github.com/cloudwego/hertz/pkg/protocol"
)

// HertzHeaderCarrier 是对 protocol.RequestHeader 的封装，使其实现 TextMapCarrier 接口
type HertzHeaderCarrier struct {
	Header *protocol.RequestHeader
}

func (c HertzHeaderCarrier) Get(key string) string {
	return string(c.Header.Get(key))
}

func (c HertzHeaderCarrier) Set(key, value string) {
	c.Header.Set(key, value)
}

func (c HertzHeaderCarrier) Keys() []string {
	keys := []string{}
	c.Header.VisitAll(func(k, _ []byte) {
		keys = append(keys, string(k))
	})
	return keys
}

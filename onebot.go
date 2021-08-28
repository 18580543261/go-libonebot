package onebot

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type OneBot struct {
	Platform string
	// ActionMux *ActionMux

	eventListenChans     []chan []byte
	eventListenChansLock sync.RWMutex

	handlers         map[string]Handler
	extendedHandlers map[string]Handler
}

func NewOneBot(platform string) *OneBot {
	if platform == "" {
		log.Warnf("没有设置 OneBot 实现平台名称, 可能导致程序行为与预期不符")
	}
	return &OneBot{
		Platform: platform,
		// ActionMux: NewActionMux(platform),

		eventListenChans:     make([]chan []byte, 0),
		eventListenChansLock: sync.RWMutex{},

		handlers:         make(map[string]Handler),
		extendedHandlers: make(map[string]Handler),
	}
}

func (ob *OneBot) startCommunicationMethods() {
	commStartHTTP("127.0.0.1", 5700, ob)
	commStartWS("127.0.0.1", 6700, ob)
	commStartHTTPWebhook("http://127.0.0.1:8080", ob)
}

func (ob *OneBot) Run() {
	ob.startCommunicationMethods()
	log.Infof("OneBot 运行中...")
	select {}
}

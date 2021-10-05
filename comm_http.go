package libonebot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type httpComm struct {
	ob               *OneBot
	latestEvents     []marshaledEvent
	latestEventsLock sync.Mutex
}

func (comm *httpComm) handleGET(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<h1>It works!</h1>"))
}

func (comm *httpComm) handle(w http.ResponseWriter, r *http.Request) {
	comm.ob.Logger.Debugf("HTTP request: %v", r)

	// reject unsupported methods
	if r.Method != "POST" && r.Method != "GET" {
		comm.ob.Logger.Warnf("动作请求只支持通过 POST 方式请求")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// handle GET requests
	if r.Method == "GET" {
		comm.handleGET(w, r)
		return
	}

	var isBinary bool
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		isBinary = false
	} else if strings.HasPrefix(contentType, "application/msgpack") {
		isBinary = true
	} else {
		// reject unsupported content types
		comm.ob.Logger.Warnf("动作请求体 MIME 类型不支持")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	// once we got the action HTTP request, we respond "200 OK"
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		comm.fail(w, RetCodeBadRequest, "动作请求体读取失败, 错误: %v", err)
		return
	}

	request, err := decodeRequest(bodyBytes, isBinary)
	if err != nil {
		comm.fail(w, RetCodeBadRequest, "动作请求解析失败, 错误: %v", err)
		return
	}

	var response Response
	if request.Action == ActionGetLatestEvents {
		// special action: get_latest_events
		response = comm.handleGetLatestEvents(&request)
	} else {
		response = comm.ob.handleRequest(&request)
	}

	respBytes, _ := comm.ob.encodeResponse(response, isBinary)
	w.Write(respBytes)
}

func (comm *httpComm) handleGetLatestEvents(r *Request) (resp Response) {
	resp.Echo = r.Echo
	w := ResponseWriter{resp: &resp}
	events := make([]AnyEvent, 0)
	// TODO: use condvar to wait until there are events
	comm.latestEventsLock.Lock()
	for _, event := range comm.latestEvents {
		events = append(events, event.raw)
	}
	comm.latestEvents = make([]marshaledEvent, 0)
	comm.latestEventsLock.Unlock()
	w.WriteData(events)
	return
}

func (comm *httpComm) fail(w http.ResponseWriter, retcode int, errFormat string, args ...interface{}) {
	err := fmt.Errorf(errFormat, args...)
	comm.ob.Logger.Warn(err)
	json.NewEncoder(w).Encode(failedResponse(retcode, err))
}

func commRunHTTP(c ConfigCommHTTP, ob *OneBot, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	ob.Logger.Infof("正在启动 HTTP (%v)...", addr)

	comm := &httpComm{
		ob:           ob,
		latestEvents: make([]marshaledEvent, 0),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", comm.handle)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ob.Logger.Errorf("HTTP (%v) 启动失败, 错误: %v", addr, err)
		}
	}()

	eventChan := ob.openEventListenChan()
	for {
		select {
		case event := <-eventChan:
			comm.latestEventsLock.Lock()
			comm.latestEvents = append(comm.latestEvents, event)
			comm.latestEventsLock.Unlock()
		case <-ctx.Done():
			ob.closeEventListenChan(eventChan)
			if err := server.Shutdown(context.TODO()); err != nil {
				ob.Logger.Errorf("HTTP (%v) 关闭失败, 错误: %v", addr, err)
			}
			ob.Logger.Infof("HTTP (%v) 已关闭", addr)
			return
		}
	}
}

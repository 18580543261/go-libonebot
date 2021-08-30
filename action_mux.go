package libonebot

import "fmt"

type ActionMux struct {
	handlers         map[string]Handler
	extendedHandlers map[string]Handler
}

func NewActionMux() *ActionMux {
	return &ActionMux{
		handlers:         make(map[string]Handler),
		extendedHandlers: make(map[string]Handler),
	}
}

func (mux *ActionMux) HandleAction(w ResponseWriter, r *Request) {
	// return "ok" if otherwise explicitly set to "failed"
	w.WriteOK()

	var handlers *map[string]Handler
	if r.Action.IsExtended {
		handlers = &mux.extendedHandlers
	} else {
		handlers = &mux.handlers
	}

	handler := (*handlers)[r.Action.Name]
	if handler == nil {
		err := fmt.Errorf("动作 `%v` 不存在", r.Action)
		w.WriteFailed(RetCodeActionNotFound, err)
		return
	}

	handler.HandleAction(w, r)
}

func (mux *ActionMux) HandleFunc(action CoreAction, handler func(ResponseWriter, *Request)) {
	mux.Handle(action, HandlerFunc(handler))
}

func (mux *ActionMux) Handle(action CoreAction, handler Handler) {
	if action.name == "" {
		panic("动作名称不能为空")
	}
	mux.handlers[action.name] = handler
}

func (mux *ActionMux) HandleFuncExtended(action string, handler func(ResponseWriter, *Request)) {
	mux.HandleExtended(action, HandlerFunc(handler))
}

func (mux *ActionMux) HandleExtended(action string, handler HandlerFunc) {
	// if the prefix is empty, then the action name starts with "_"
	mux.extendedHandlers[action] = handler
}

package natskit

// MsgHandlerMid is function designed to run some code before/after given MsgHandler
type MsgHandlerMid func(MsgHandler) MsgHandler

// wrapMiddleware creates a new final MsgHandler by wrapping middlewares around the original one
func wrapMiddleware(handler MsgHandler, mw []MsgHandlerMid) MsgHandler {
	if len(mw) == 0 {
		return handler
	}

	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}

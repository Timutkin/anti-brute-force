package handler

type WhiteListHandler struct {
	ListHandler
}

func NewWhiteListHandler(handler ListHandler) WhiteListHandler {
	return WhiteListHandler{handler}
}

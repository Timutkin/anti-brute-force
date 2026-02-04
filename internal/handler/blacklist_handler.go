package handler

type BlackListHandler struct {
	ListHandler
}

func NewBlackListHandler(handler ListHandler) BlackListHandler {
	return BlackListHandler{handler}
}

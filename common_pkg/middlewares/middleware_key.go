package common_middlewares

type contextKey string

const (
	CtxRequestID           contextKey = "requestId"
	CtxUserEmail           contextKey = "email"
)

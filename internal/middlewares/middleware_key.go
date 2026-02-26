package middlewares

type contextKey string

const (
	CtxRegistrationPayload contextKey = "registration_payload"
	CtxUpdatePayload       contextKey = "update_payload"
)

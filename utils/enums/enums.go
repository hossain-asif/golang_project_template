package enums

// middleware related enums
type contextKey string

const (
	CtxRegistrationPayload contextKey = "registration_payload"
	CtxUpdatePayload       contextKey = "update_payload"
	CtxRequestID           contextKey = "requestId"
	CtxUserEmail           contextKey = "email"
	CtxUserSlug            contextKey = "user_slug"
	CtxLoginPayload        contextKey = "login_payload"
)

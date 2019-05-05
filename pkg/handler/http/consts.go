package http

const (
	HeaderAuthorization             = "Authorization"
	HeaderAuthorizationBearerPrefix = "bearer "
	keyAPIKey                       = "x-api-key"

	ctxKeyLog    = contextKey("log")
	ctxKeyClaims = contextKey("claims")
)

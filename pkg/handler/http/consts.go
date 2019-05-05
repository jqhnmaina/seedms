package http

const (
	HeaderAuthorization             = "Authorization"
	HeaderAuthorizationBearerPrefix = "bearer "

	ctxKeyLog    = contextKey("log")
	ctxKeyClaims = contextKey("claims")
)

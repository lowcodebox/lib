package models

const (
	headerXRequestID      = "X-Request-ID"
	headerXUserID         = "X-User-ID"
	headerXRequestUnit    = "X-Request-Unit"
	headerXRequestService = "X-Request-Service"

	requestIDField = "request-id"
	userIDField    = "user-id"
	serviceIDField = "service-id"
	configIDField  = "config-id"
)

var ProxiedHeaders = map[string]string{
	requestIDField: headerXRequestID,
	userIDField:    headerXUserID,
}

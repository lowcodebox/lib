package models

const (
	RequestIDField = "request-id"
	UserIDField    = "user-id"
	UserIPField    = "user-ip"
	ClientField    = "client"
	ServiceIDField = "service-id"
	ConfigIDField  = "config-id"
	DCField        = "dc"

	HeaderXRequestID      = "X-Request-ID"
	HeaderXServiceKey     = "X-Service-Key"
	HeaderXAuthKey        = "X-Auth-Key"
	HeaderXUserID         = "X-User-ID"
	HeaderXUserIP         = "X-User-IP"
	HeaderXRequestUnit    = "X-Request-Unit"
	HeaderXRequestService = "X-Request-Service"
	HeaderXServiceClient  = "X-Service-Client"
	HeaderXDC             = "X-DC"
)

var ProxiedHeaders = map[string]string{
	RequestIDField: HeaderXRequestID,
	UserIDField:    HeaderXUserID,
	ClientField:    HeaderXServiceClient,
	UserIPField:    HeaderXUserIP,
	DCField:        HeaderXDC,
}

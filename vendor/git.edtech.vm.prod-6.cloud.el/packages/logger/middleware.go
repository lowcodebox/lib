package logger

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/segmentio/ksuid"
	_ "github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"git.edtech.vm.prod-6.cloud.el/packages/logger/types"
)

const (
	headerXRequestID      = "X-Request-ID"
	headerXUserID         = "X-User-ID"
	headerXRequestUnit    = "X-Request-Unit"
	headerXRequestService = "X-Request-Service"

	requestIDKey        = "request-id"
	ServiceTypeKey      = "service-type"
	ServiceIDKey        = "service-id"
	ConfigIDKey         = "config-id"
	requestUserAgentKey = "user-agent"
	requestUnitKey      = "unit-key"
	requestServiceKey   = "service-key"
	requestBodyKey      = "request-body"
	responseBodyKey     = "response-body"
	methodKey           = "method"
	remoteAddrKey       = "remote-addr"
	userAgentKey        = "user-agent"
	durationKey         = "duration"
	urlKey              = "url"
	statusCodeKey       = "status-code"
	headerExpiresAtKey  = "header_expires_at"
	requestLengthKey    = "request-length"
)

type statusCodeWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusCodeWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusCodeWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = 200
	}

	return w.ResponseWriter.Write(b)
}

func (w *statusCodeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok || hj == nil {
		return nil, nil, fmt.Errorf("http.Hijacker interface is not implemented in given response writer")
	}

	return hj.Hijack()
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(headerXRequestID)
		if requestID == "" {
			requestID = ksuid.New().String()
			r.Header.Add(headerXRequestID, requestID) // keep compatibility with old logger
		}
		w.Header().Set(headerXRequestID, requestID)
		ctx := SetRequestIDCtx(r.Context(), requestID)

		userID := r.Header.Get(headerXUserID)
		if userID != "" {
			w.Header().Set(headerXUserID, userID)
			ctx = SetUserIDCtx(ctx, userID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HttpClientHeaders устанавливает заголовки реквеста из контекста и headers
func HttpClientHeaders(ctx context.Context, req *http.Request, headers map[string]string) {
	if req == nil {
		return
	}

	for ctxField, headerField := range models.ProxiedHeaders {
		if value := GetFieldCtx(ctx, ctxField); value != "" {
			req.Header.Add(headerField, value)
		}
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
}

// HTTPMiddleWareParams параметры логирования, при пустых параметрах HTTPMiddlewareWithParams
// будет отрабатывать, как HTTPMiddleware.
type HTTPMiddleWareParams struct {
	NeedToLogIncomingPostRequest bool
	NeedToLogResponse            bool
	NeedToLogPanic               bool
}

type Middleware interface {
	HTTPMiddlewareWithParams(next http.Handler) http.Handler
}

type middleware struct {
	params *HTTPMiddleWareParams
}

func NewHTTPMiddleware(params *HTTPMiddleWareParams) Middleware {
	return &middleware{params: params}
}

// HTTPMiddlewareWithParams http middleware для записи логов запросов,
// параметры логирования определяются HTTPMiddleWareParams.
func (m *middleware) HTTPMiddlewareWithParams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.params.NeedToLogPanic {
			defer func() {
				if rr := recover(); rr != nil {
					stack := debug.Stack()
					pc, file, line, _ := runtime.Caller(4)
					Logger(r.Context()).Warn("panic happened",
						zap.String("recover-error", fmt.Sprint(rr)),
						zap.String("file", file),
						zap.Int("line", line),
						zap.String("function", runtime.FuncForPC(pc).Name()),
						zap.ByteString("stack", stack),
						zap.String("stacktrace", string(debug.Stack())),
						zap.Int(statusCodeKey, 500),
						types.URL("url", r.RequestURI),
						zap.String("method", r.Method),
						zap.String("remote-addr", r.RemoteAddr),
						zap.String("user-agent", r.Header.Get("User-Agent")),
						zap.String("header_expires_at", r.Header.Get("X-Expires-At")))
					panic(r)
				}
			}()
		}

		var timeStart time.Time

		if m.params.NeedToLogResponse {
			timeStart = time.Now()
		}

		requestID := r.Header.Get(headerXRequestID)
		if requestID == "" {
			requestID = ksuid.New().String()
			r.Header.Add(headerXRequestID, requestID) // keep compatibility with old logger
		}

		w.Header().Set(headerXRequestID, requestID)
		ctx := SetRequestIDCtx(r.Context(), requestID)

		userID := r.Header.Get(headerXUserID)
		if userID != "" {
			w.Header().Set(headerXUserID, userID)
			ctx = SetUserIDCtx(ctx, userID)
		}

		if m.params.NeedToLogIncomingPostRequest && isPostWithContent(r) {
			Logger(r.Context()).Info("incoming POST request",
				types.URL("url", r.RequestURI),
				zap.String(requestUnitKey, r.Header.Get(headerXRequestUnit)),
				zap.String(requestServiceKey, r.Header.Get(headerXRequestService)),
				zap.String(requestBodyKey, getRequestBodyCopy(r)),
				zap.Int64("request-length", r.ContentLength),
				zap.String("method", r.Method),
				zap.String("remote-addr", r.RemoteAddr),
				zap.String(requestUserAgentKey, r.Header.Get("User-Agent")),
				zap.String("header_expires_at", r.Header.Get("X-Expires-At")))
		}

		sw := &statusCodeWriter{ResponseWriter: w}
		*r = *r.WithContext(ctx)
		next.ServeHTTP(sw, r)

		if m.params.NeedToLogResponse {
			Logger(r.Context()).Info("request completed",
				zap.Float64("timing", time.Since(timeStart).Seconds()),
				zap.Int(statusCodeKey, sw.statusCode),
				types.URL("url", r.RequestURI),
				zap.String(requestUnitKey, r.Header.Get(headerXRequestUnit)),
				zap.String(requestServiceKey, r.Header.Get(headerXRequestService)),
				zap.String(requestBodyKey, getRequestBodyCopy(r)),
				zap.String("method", r.Method),
				zap.String("remote-addr", r.RemoteAddr),
				zap.String("user-agent", r.Header.Get("User-Agent")),
				zap.String("header_expires_at", r.Header.Get("X-Expires-At")))
		}
	})
}

type loggerFormat struct {
	logRequestFunc  func(req *http.Request)
	logResponseFunc func(context.Context, *http.Response, time.Time)
}

func HTTPClientLogging(hc *http.Client, options ...Option) {
	lf := &loggerFormat{
		logRequestFunc:  defaultLogRequest,
		logResponseFunc: defaultLogResponse,
	}
	for _, option := range options {
		lf = option(lf)
	}

	if hc.Transport == nil {
		hc.Transport = http.DefaultTransport
	}

	hc.Transport = &loggerTransport{
		transport:       hc.Transport,
		logRequestFunc:  lf.logRequestFunc,
		logResponseFunc: lf.logResponseFunc,
	}
}

// loggerTransport implements http.RoundTripper.
// When set as Transport of http.Client, it executes HTTP requests with logging.
// No field is mandatory.
type loggerTransport struct {
	transport       http.RoundTripper
	logRequestFunc  func(req *http.Request)
	logResponseFunc func(context.Context, *http.Response, time.Time)
}

// Used if loggerTransport.logRequestFunc is not set.
var defaultLogRequest = func(r *http.Request) {
	Logger(r.Context()).Info("external request",
		zap.String(methodKey, r.Method),
		types.URL(urlKey, r.URL.String()),
		types.JSON(requestBodyKey, getRequestBodyCopy(r)),
		zap.Int64(requestLengthKey, r.ContentLength))
}

// Used if loggerTransport.logResponseFunc is not set.
var defaultLogResponse = func(ctx context.Context, r *http.Response, begin time.Time) {
	Logger(ctx).Info("external response",
		zap.String(methodKey, r.Request.Method),
		zap.Int(statusCodeKey, r.StatusCode),
		types.URL(urlKey, r.Request.URL.String()),
		types.JSON(responseBodyKey, getResponseBodyCopy(r)),
		zap.Float64(durationKey, time.Since(begin).Seconds()),
	)
}

// ExternalFormRequestLogger hides sensitive data from request.
// hides from url.
// hides from form values.
func ExternalFormRequestLogger() func(r *http.Request) {
	return func(r *http.Request) {
		Logger(r.Context()).Info("external request",
			zap.String(methodKey, r.Method),
			types.URL(urlKey, r.URL.String()),
			types.URL(requestBodyKey, getRequestBodyCopy(r)),
			zap.Int64(requestLengthKey, r.ContentLength))
	}
}

// RoundTrip is the core part of this module and implements http.RoundTripper.
// Executes HTTP request with request/response logging.
func (lt *loggerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	lt.logRequest(req)

	begin := time.Now()

	resp, err := lt.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	lt.logResponse(req.Context(), resp, begin)

	return resp, err
}

func (lt *loggerTransport) logRequest(req *http.Request) {
	lt.logRequestFunc(req)
}

func (lt *loggerTransport) logResponse(ctx context.Context, resp *http.Response, begin time.Time) {
	lt.logResponseFunc(ctx, resp, begin)
}

func GRPCUnaryServerInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	var requestID, userID string

	if headers, ok := metadata.FromIncomingContext(ctx); ok {
		if header, ok := headers["x-request-id"]; ok && len(header) > 0 {
			requestID = header[0]
		}
		if header, ok := headers["x-user-id"]; ok && len(header) > 0 {
			userID = header[0]
		}
	}

	if requestID == "" {
		requestID = ksuid.New().String()
	}

	ctx = AddToGRPCHeader(ctx, "x-request-id", requestID)
	ctx = SetRequestIDCtx(ctx, requestID)

	if userID != "" {
		ctx = AddToGRPCHeader(ctx, "x-user-id", userID)
		ctx = SetRequestIDCtx(ctx, userID)
	}

	return handler(ctx, req)
}

func GRPCUnaryClientInterceptor(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	var requestID, userID string

	requestID = GetRequestIDCtx(ctx)
	userID = GetUserIDCtx(ctx)

	if requestID == "" {
		requestID = ksuid.New().String()
	}

	ctx = AddToGRPCHeader(ctx, "x-request-id", requestID)
	if userID != "" {
		ctx = AddToGRPCHeader(ctx, "x-user-id", userID)
	}

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func getRequestBodyCopy(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	var reqBody []byte

	reqBody, _ = io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	return string(reqBody)
}

func isPostWithContent(r *http.Request) bool {
	return r.Method == http.MethodPost && r.ContentLength != 0
}

func getResponseBodyCopy(r *http.Response) string {
	if r.Body == nil {
		return ""
	}

	var reqBody []byte

	reqBody, _ = io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

	return string(reqBody)
}

func GRPCLoggingInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (
	interface{}, error,
) {
	begin := time.Now()

	logMethodBefore(ctx, info.FullMethod, req)

	res, err := handler(ctx, req)

	logMethodAfter(ctx, info.FullMethod, begin, res, err)

	return res, err
}

func logMethodAfter(ctx context.Context, path string, begin time.Time, resp interface{}, err error) {
	parts := strings.Split(path, "/")
	methodName := parts[len(parts)-1]

	respRaw, respErr := json.Marshal(resp)
	if respErr != nil {
		Logger(ctx).Warn("logMethodAfter err:", zap.Error(respErr))
	}

	fields := []zap.Field{
		zap.String("method", methodName),
		types.JSON("response", string(respRaw)),
		zap.Float64("duration", time.Since(begin).Seconds()),
	}

	if err != nil {
		fields = append(fields, zap.String("error", err.Error()))
	}

	Logger(ctx).Info("finished", fields...)
}

func logMethodBefore(ctx context.Context, path string, req interface{}) {
	parts := strings.Split(path, "/")
	methodName := parts[len(parts)-1]

	reqRaw, err := json.Marshal(req)
	if err != nil {
		Logger(ctx).Warn("logMethodBefore err:", zap.Error(err))
	}

	Logger(ctx).Info("enter", zap.String("method", methodName), types.JSON("request", string(reqRaw)))
}

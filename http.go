package lib

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/labstack/gommon/color"
)

// ResponseWrapper Обертка над ResponseWriter для сбора статус кодов
type ResponseWrapper struct {
	http.ResponseWriter

	Code int
}

func (r *ResponseWrapper) WriteHeader(statusCode int) {
	r.Code = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type curlParam struct {
	method          string
	url             string
	body            string
	response        interface{}
	headers         map[string]string
	cookies         []*http.Cookie
	disableRedirect bool
	timeout         time.Duration
}

type curlResp struct {
	result     interface{}
	headers    http.Header
	cookies    []*http.Cookie
	statusCode int
}

const clientHttpTimeout = 60 * time.Second

var (
	reCrLf = regexp.MustCompile(`[\r\n]+`)
	rePort = regexp.MustCompile(`:\d+$`)

	errUnauthorized = errors.New("unauthorized")
	errTokenInvalid = errors.New("token is not valid")
)

// Curl всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func Curl(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, status int, err error) {
	r, _, _, status, err := curlEngine(ctx, method, urlc, bodyJSON, response, headers, cookies)
	return r, status, err
}

func CurlCookies(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, resp_cookies []*http.Cookie, status int, err error) {
	r, _, cookies, status, err := curlEngine(ctx, method, urlc, bodyJSON, response, headers, cookies)
	return r, cookies, status, err
}

// CurlV2 - версия curl, которая возвращает заголовки и куки
func CurlV2(
	ctx context.Context,
	method, urlc, bodyJSON string,
	response interface{},
	headers map[string]string,
	cookies []*http.Cookie,
	timeout time.Duration,
	disableRedirect bool) (result interface{}, respHeaders http.Header, respCookies []*http.Cookie, status int, err error) {
	in := curlParam{
		method:          method,
		url:             urlc,
		body:            bodyJSON,
		response:        response,
		headers:         headers,
		cookies:         cookies,
		disableRedirect: disableRedirect,
		timeout:         timeout,
	}

	resp, err := curlEngineV2(ctx, in)
	return resp.result, resp.headers, resp.cookies, resp.statusCode, err
}

// curlEngineV2 - версия curl, которая возвращает заголовки, куки и не следует redirect
// Также в req *http.Request пробрасывается ctx
func curlEngineV2(ctx context.Context, args curlParam) (result curlResp, err error) {
	result.result = ""
	var mapValues map[string]string
	var req *http.Request
	var checkRedirect func(r *http.Request, via []*http.Request) error

	if args.timeout == 0 {
		args.timeout = clientHttpTimeout
	}

	if args.disableRedirect {
		checkRedirect = func(r *http.Request, via []*http.Request) error {
			// Возвращаем ошибку, чтобы остановить следование редиректу
			return http.ErrUseLastResponse
		}
	}

	client := &http.Client{
		Timeout:       args.timeout,
		CheckRedirect: checkRedirect,
	}

	dialer := net.Dialer{
		Timeout: args.timeout,
	}

	//nolint:gosec
	tlsConfig := tls.Config{
		InsecureSkipVerify: true, // ignore/skip expired SSL certificates
	}

	transCfg := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: args.timeout / 5,
		TLSClientConfig:     &tlsConfig,
	}

	client.Transport = transCfg

	if args.method == "" {
		args.method = http.MethodGet
	}

	args.method = strings.TrimSpace(args.method)
	values := url.Values{}
	actionType := ""

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(args.url, "?")

	args.body = reCrLf.ReplaceAllString(args.body, "")

	if args.method == "JSONTOGET" && args.body != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if args.method == "JSONTOPOST" && args.body != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		err = json.Unmarshal([]byte(args.body), &mapValues)
		if err != nil {
			return result, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %w", args.body, err)
		}

		for k, v := range mapValues {
			values.Set(k, v)
		}
		uri, _ := url.Parse(args.url)
		uri.RawQuery = values.Encode()
		args.url = uri.String()
		req, err = http.NewRequestWithContext(ctx, "GET", args.url, strings.NewReader(args.body))

	case "JSONTOPOST": // преобразуем параметры в json в тело запроса
		err = json.Unmarshal([]byte(args.body), &mapValues)
		if err != nil {
			return result, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %w", args.body, err)
		}

		for k, v := range mapValues {
			values.Set(k, v)
		}
		req, err = http.NewRequestWithContext(ctx, "POST", args.url, strings.NewReader(values.Encode()))
		req.PostForm = values
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	default:
		req, err = http.NewRequestWithContext(ctx, args.method, args.url, strings.NewReader(args.body))
	}

	if err != nil {
		return result, fmt.Errorf("error NewRequest in lib.Curl, err: %w", err)
	}

	// дополняем переданными заголовками
	httpClientHeaders(ctx, req, args.headers)

	// дополянем куками назначенными для данного запроса
	if args.cookies != nil {
		for _, v := range args.cookies {
			req.AddCookie(v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		result.statusCode = http.StatusBadRequest
		return
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		result.statusCode = resp.StatusCode
		return
	}

	// возвращаем объект ответа, если передано - в какой объект класть результат
	// НА ОШИБКУ НЕ ПРОВЕРЯТЬ!!!!!!
	if args.response != nil {
		_ = json.Unmarshal(responseData, &args.response)
	}

	result.result = string(responseData)
	result.headers = resp.Header
	result.cookies = resp.Cookies()
	result.statusCode = resp.StatusCode

	return
}

func curlEngine(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, respHeaders http.Header, resp_cookies []*http.Cookie, status int, err error) {
	var mapValues map[string]string
	var req *http.Request
	var skipTLSVerify = true

	client := &http.Client{
		Timeout: clientHttpTimeout,
	}

	dialer := net.Dialer{
		Timeout: clientHttpTimeout,
	}

	//nolint:gosec
	tlsConfig := tls.Config{
		InsecureSkipVerify: skipTLSVerify, // ignore expired SSL certificates
	}

	transCfg := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: clientHttpTimeout / 5,
		TLSClientConfig:     &tlsConfig,
	}

	client.Transport = transCfg

	if method == "" {
		method = "POST"
	}

	method = strings.Trim(method, " ")
	values := url.Values{}
	actionType := ""

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(urlc, "?")

	bodyJSON = reCrLf.ReplaceAllString(bodyJSON, "")

	if method == "JSONTOGET" && bodyJSON != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if method == "JSONTOPOST" && bodyJSON != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		err = json.Unmarshal([]byte(bodyJSON), &mapValues)
		if err != nil {
			return nil, nil, nil, status, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
		}

		for k, v := range mapValues {
			values.Set(k, v)
		}
		uri, _ := url.Parse(urlc)
		uri.RawQuery = values.Encode()
		urlc = uri.String()
		req, err = http.NewRequest("GET", urlc, strings.NewReader(bodyJSON))

	case "JSONTOPOST": // преобразуем параметры в json в тело запроса
		err = json.Unmarshal([]byte(bodyJSON), &mapValues)
		if err != nil {
			return nil, nil, nil, status, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
		}

		for k, v := range mapValues {
			values.Set(k, v)
		}
		req, err = http.NewRequest("POST", urlc, strings.NewReader(values.Encode()))
		req.PostForm = values
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	default:
		req, err = http.NewRequest(method, urlc, strings.NewReader(bodyJSON))
	}

	if err != nil {
		return nil, nil, nil, status, fmt.Errorf("error NewRequest in lib.Curl, err: %w", err)
	}

	// дополняем переданными заголовками
	httpClientHeaders(ctx, req, headers)

	// дополянем куками назначенными для данного запроса
	if cookies != nil {
		for _, v := range cookies {
			req.AddCookie(v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		//fmt.Println("Error request: method:", method, ", url:", urlc, ", bodyJSON:", bodyJSON, "err:", err)
		return "", nil, nil, http.StatusBadRequest, err
	} else {
		defer resp.Body.Close()
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, nil, resp.StatusCode, err
	}
	responseString := string(responseData)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	// НА ОШИБКУ НЕ ПРОВЕРЯТЬ!!!!!!
	if response != nil {
		_ = json.Unmarshal([]byte(responseString), &response)
	}

	//// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	////json.Unmarshal([]byte(responseString), &result)
	//if resp.statusCode != 200 {
	//	err = fmt.Errorf("request is not success. request: %s, status: %s, method: %s, req: %+v, response: %s", urlc, resp.statusCode, method, req, responseString)
	//}
	//
	status = resp.StatusCode

	return responseString, resp.Header, resp.Cookies(), status, err
}

func AddressProxy(addressProxy, interval string) (port string, err error) {
	var res interface{}
	fail := color.Red("[Fail]")
	urlProxy := ""

	// если автоматическая настройка портов
	if addressProxy != "" && interval != "" {
		if addressProxy[len(addressProxy)-1:] != "/" {
			addressProxy = addressProxy + "/"
		}

		var portDataAPI models.Response
		// запрашиваем порт у указанного прокси-сервера
		urlProxy = addressProxy + "port?interval=" + interval

		res, _, err = Curl(context.Background(), "GET", urlProxy, "", &portDataAPI, map[string]string{}, nil)
		if err != nil {
			return "", err
		}
		port = fmt.Sprint(portDataAPI.Data)
	}

	if port == "" {
		err = fmt.Errorf("Port APP-service is null. Servive not running. (urlProxy: %s, response: %+v)", urlProxy, res)
		fmt.Print(fail, err, "\n")
	}

	return port, err
}

func ClearSlash(url string) (result string) {
	if len(url) == 0 {
		return ""
	}
	// удаляем слеш сзади
	lastSleshF := url[len(url)-1:]
	if lastSleshF == "/" {
		url = url[:len(url)-1]
	}

	// удаляем слеш спереди
	lastSleshS := url[0:1]
	if lastSleshS == "/" {
		url = url[1:len(url)]
	}

	return url
}

func PortResolver(port string) (status bool) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}

	ln.Close()
	return true
}

// ProxyPort свободный порт от прокси с проверкой доступности на локальной машине
// если занято - ретраим согласно заданным параметрам
func ProxyPort(addressProxy, interval string, maxCountRetries int, timeRetries time.Duration) (port string, err error) {
	port, err = Retrier(maxCountRetries, timeRetries, true, func() (string, error) {
		port, err = AddressProxy(addressProxy, interval)
		if err != nil {
			return "", err
		}

		status := PortResolver(port)
		if status {
			return port, nil
		}

		return "", fmt.Errorf("listen tcp :%s. address already in use", port)
	})

	return port, err
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Forwarded-For")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}

	if strings.Contains(IPAddress, ":") {
		// заменяем так, потому что в IPv6 присутствует «:»
		return rePort.ReplaceAllString(IPAddress, "")
	}

	return IPAddress
}

// httpClientHeaders устанавливает заголовки реквеста из контекста и headers
func httpClientHeaders(ctx context.Context, req *http.Request, headers map[string]string) {
	if req == nil {
		return
	}
	for ctxField, headerField := range models.ProxiedHeaders {
		if value := getFieldCtx(ctx, ctxField); value != "" {
			req.Header.Add(headerField, value)
		}
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
}

func getFieldCtx(ctx context.Context, name string) string {
	if ctx == nil {
		return ""
	}
	nameKey := "logger." + name
	a := ctx.Value(nameKey)
	if a == nil {
		return ""
	}
	str, ok := a.(string)
	if !ok {
		return fmt.Sprint(a)
	}

	return str
}

func ExtractNameVersionString(target, defaultName, defaultVersion string) (name, version, host string, err error) {
	path := target
	if len(path) > 1 && path[0] == '/' {
		path = path[1:]
	}
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		return defaultName, defaultVersion, "", nil
	}
	name, version = tmp[0], tmp[1]
	target = "/" + strings.Join(tmp[2:], "/")
	return name, version, host, nil
}

func CheckIntranet(req *http.Request) bool {
	ip := ReadUserIP(req)
	switch {
	case ip == "127.0.0.1",
		ip == "[::1]",
		strings.HasPrefix(ip, "10."),
		strings.HasPrefix(ip, "192.168."),
		strings.HasPrefix(ip, "172.17."):

		return true
	}

	return false
}

func getServiceKey(r *http.Request) string {
	serviceKey := r.Header.Get(models.HeaderXServiceKey)
	if serviceKey == "" {
		serviceKeyCookie, err := r.Cookie(models.HeaderXServiceKey)
		if err == nil {
			serviceKey = serviceKeyCookie.Value
		}
	}
	return serviceKey
}

func MiddlewareXServiceKey(project, service, projectKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			ctx := r.Context()

			defer func() {
				if err != nil {
					_ = ResponseJSON(w, nil, "Unauthorized", err, nil)
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			serviceKey := getServiceKey(r)

			if serviceKey == "" {
				err = errUnauthorized
				return
			}

			valid, _ := CheckXServiceKey(project+"/"+service, []byte(projectKey), serviceKey)
			if !valid {
				err = errTokenInvalid
			}
		})
	}
}

// MiddlewareValidateUri проверяет токен на доступ к пути
// Инвертирует логику WhiteUri = "". IsValidURI по умолчанию пропускает. Эта не будет пропускать
func MiddlewareValidateUri(projectKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получить ключ
			xsKey, err := decodeServiceKey([]byte(projectKey), getServiceKey(r))
			if err != nil {
				_ = ResponseJSON(w, nil, "Unauthorized", errTokenInvalid, nil)
				return
			}

			// Получить список доступа
			if xsKey.WhiteURI == "" {
				_ = ResponseJSON(w, nil, "Unauthorized", errUnauthorized, nil)
				return
			}

			// Проверить список доступа
			valid := isValidUri(xsKey, r.URL.Path)
			if !valid {
				_ = ResponseJSON(w, nil, "Unauthorized", errUnauthorized, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

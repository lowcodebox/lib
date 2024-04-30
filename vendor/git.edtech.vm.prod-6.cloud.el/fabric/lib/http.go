package lib

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/labstack/gommon/color"
)

const clientHttpTimeout = 60 * time.Second

var reCrLf = regexp.MustCompile(`[\r\n]+`)

// Curl всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func Curl(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, err error) {
	r, _, err := curl_engine(ctx, method, urlc, bodyJSON, response, headers, cookies)
	return r, err
}

func CurlCookies(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, resp_cookies []*http.Cookie, err error) {
	return curl_engine(ctx, method, urlc, bodyJSON, response, headers, cookies)
}

func curl_engine(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, resp_cookies []*http.Cookie, err error) {
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
			return nil, nil, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
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
			return nil, nil, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
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
		return "", nil, err
	} else {
		defer resp.Body.Close()
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	responseString := string(responseData)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	// НА ОШИБКУ НЕ ПРОВЕРЯТЬ!!!!!!
	if response != nil {
		json.Unmarshal([]byte(responseString), &response)
	}

	// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	//json.Unmarshal([]byte(responseString), &result)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("request is not success. request: %s, status: %s, method: %s, req: %+v, response: %s", urlc, resp.Status, method, req, responseString)
	}

	return responseString, resp.Cookies(), err
}

func AddressProxy(addressProxy, interval string) (port string, err error) {
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
		_, err := Curl(context.Background(), "GET", urlProxy, "", &portDataAPI, map[string]string{}, nil)
		if err != nil {
			return "", err
		}
		port = fmt.Sprint(portDataAPI.Data)
	}

	if port == "" {
		err = fmt.Errorf("%s", "Port APP-service is null. Servive not running.")
		fmt.Print(fail, " Port APP-service is null. Servive not running.\n")
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
	requestID, ok := a.(string)
	if !ok {
		return ""
	}

	return requestID
}

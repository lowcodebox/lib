package lib

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
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

const clientHttpTimeout = 60 * time.Second

var (
	reCrLf = regexp.MustCompile(`[\r\n]+`)
	rePort = regexp.MustCompile(`:\d+$`)

	errUnauthorized = errors.New("unauthorized")
	errTokenInvalid = errors.New("token is not valid")
)

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

func MiddlewareXServiceKey(name, version, projectKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var serviceKey string
			var err error

			ctx := r.Context()

			defer func() {
				if err != nil {
					_ = ResponseJSON(w, nil, "Unauthorized", err, nil)
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			// он приватный - проверяем на валидность токена
			serviceKeyHeader := r.Header.Get(models.HeaderXServiceKey)
			if serviceKeyHeader != "" {
				serviceKey = serviceKeyHeader
			} else {
				serviceKeyCookie, err := r.Cookie(models.HeaderXServiceKey)
				if err == nil {
					serviceKey = serviceKeyCookie.Value
				}
			}

			if serviceKey == "" {
				err = errUnauthorized
				return
			}

			valid, _ := CheckXServiceKey(name+"/"+version, []byte(projectKey), serviceKey)
			if !valid {
				err = errTokenInvalid
			}
		})
	}
}

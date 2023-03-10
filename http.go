package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/models"
	"github.com/labstack/gommon/color"
)

// Curl всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func Curl(method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, err error) {
	var mapValues map[string]string
	var req *http.Request
	client := &http.Client{}
	client.Timeout = 3 * time.Second

	if method == "" {
		method = "POST"
	}

	method = strings.Trim(method, " ")
	values := url.Values{}
	actionType := ""

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(urlc, "?")

	bodyJSON = strings.Replace(bodyJSON, "  ", "", -1)
	err = json.Unmarshal([]byte(bodyJSON), &mapValues)

	if method == "JSONTOGET" && bodyJSON != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if method == "JSONTOPOST" && bodyJSON != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			uri, _ := url.Parse(urlc)
			uri.RawQuery = values.Encode()
			urlc = uri.String()
			req, err = http.NewRequest("GET", urlc, strings.NewReader(bodyJSON))
		} else {
			fmt.Println("Error! Fail parsed bodyJSON from GET Curl: ", err)
		}
	case "JSONTOPOST": // преобразуем параметры в json в тело запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			req, err = http.NewRequest("POST", urlc, strings.NewReader(values.Encode()))
			req.PostForm = values
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			fmt.Println("Error! Fail parsed bodyJSON to POST: ", err)
		}
	default:
		req, err = http.NewRequest(method, urlc, strings.NewReader(bodyJSON))
	}

	if err != nil {
		return "", err
	}

	// дополняем переданными заголовками
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	// дополянем куками назначенными для данного запроса
	if cookies != nil {
		for _, v := range cookies {
			req.AddCookie(v)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error request: method:", method, ", url:", urlc, ", bodyJSON:", bodyJSON, "err:", err)
		return "", err
	} else {
		defer resp.Body.Close()
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseString := string(responseData)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	if response != nil {
		json.Unmarshal([]byte(responseString), &response)
	}

	// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	//json.Unmarshal([]byte(responseString), &result)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("request is not success. status: %s", resp.Status)
	}

	return responseString, err
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
		_, err := Curl("GET", urlProxy, "", &portDataAPI, map[string]string{}, nil)
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

func Ping(version, project, domain, port, grpc, metric, uid, state string, replicas int) (result []models.Pong, err error) {
	var r = []models.Pong{}

	name, version := ValidateNameVersion(project, version, domain)
	pg, _ := strconv.Atoi(port)
	gp, _ := strconv.Atoi(grpc)
	mp, _ := strconv.Atoi(metric)
	pid := strconv.Itoa(os.Getpid())
	//state, _ := json.Marshal(s.metrics.Get())

	pong := models.Pong{}
	pong.Uid = uid
	pong.Name = name
	pong.Version = version
	pong.Status = "run"
	pong.Port = pg
	pong.Pid = pid
	pong.State = state
	pong.Replicas = replicas
	pong.Https = false
	pong.Grpc = gp
	pong.Metric = mp

	r = append(r, pong)

	return r, err
}

package curl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type request struct {
	method, url, payload string
	response             interface{}
	headers              map[string]string
	cookies              []*http.Cookie

	client http.Client
}

type Builder interface {
	Method(value string) Builder
	Url(value string) Builder
	Payload(value string) Builder
	MapToObj(value interface{}) Builder
	Headers(value map[string]string) Builder
	Cookies(value []*http.Cookie) Builder

	Do(ctx context.Context) (result interface{}, err error)
}

func (b *request) Method(value string) Builder {
	b.method = value
	return b
}

func (b *request) Url(value string) Builder {
	b.url = value
	return b
}

func (b *request) Payload(value string) Builder {
	b.payload = value
	return b
}

func (b *request) MapToObj(value interface{}) Builder {
	b.response = value
	return b
}

func (b *request) Headers(value map[string]string) Builder {
	b.headers = value
	return b
}

func (b *request) Cookies(value []*http.Cookie) Builder {
	b.cookies = value
	return b
}

// Do всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func (r *request) Do(ctx context.Context) (result interface{}, err error) {
	var mapValues map[string]string
	var req *http.Request

	if r.method == "" {
		r.method = http.MethodPost
	}

	if r.url == "" {
		return nil, fmt.Errorf("error do request. param URL is empty")
	}

	values := url.Values{}
	actionType := ""

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(r.url, "?")

	r.payload = strings.Replace(r.payload, "  ", "", -1)
	err = json.Unmarshal([]byte(r.payload), &mapValues)

	if r.method == "JSONTOGET" && r.payload != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if r.method == "JSONTOPOST" && r.payload != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			uri, _ := url.Parse(r.url)
			uri.RawQuery = values.Encode()
			r.url = uri.String()
			req, err = http.NewRequest("GET", r.url, strings.NewReader(r.payload))
			if err != nil {
				return "", fmt.Errorf("error do request, err: %s", err)
			}
		} else {
			fmt.Println("Error! Fail parsed bodyJSON from GET Curl: ", err)
		}
	case "JSONTOPOST": // преобразуем параметры в json в тело запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			req, err = http.NewRequest("POST", r.url, strings.NewReader(values.Encode()))
			if err != nil {
				return "", fmt.Errorf("error do request, err: %s", err)
			}

			req.PostForm = values
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			fmt.Println("Error! Fail parsed bodyJSON to POST: ", err)
		}
	default:
		req, err = http.NewRequest(r.method, r.url, strings.NewReader(r.payload))
		if err != nil {
			return "", fmt.Errorf("error do request, err: %s", err)
		}
	}

	// дополняем переданными заголовками
	if len(r.headers) > 0 {
		for k, v := range r.headers {
			req.Header.Add(k, v)
		}
	}

	// дополянем куками назначенными для данного запроса
	if r.cookies != nil {
		for _, v := range r.cookies {
			req.AddCookie(v)
		}
	}

	if ctx == nil {
		ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	}

	req = req.WithContext(ctx)
	req.Close = true

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error do request, err: %s", err)
	}
	defer resp.Body.Close()

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseString := string(responseData)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	if r.response != nil {
		json.Unmarshal([]byte(responseString), &r.response)
	}

	// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	//json.Unmarshal([]byte(responseString), &result)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("request is not success (do request). request: %s, status: %s", r.url, resp.Status)
	}

	return responseString, err
}

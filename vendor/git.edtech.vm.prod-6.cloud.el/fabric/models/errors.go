package models

import "fmt"

type ErrorClient struct {
	ServiceName string
	Url         string
	Path        string
	Status      int
	Err         error
}

func (e ErrorClient) Error() string {
	return fmt.Sprintf("error request to %s (%s). url: %s status: %d, err: %s", e.ServiceName, e.Path, e.Url, e.Status, e.Err)
}

func (e ErrorClient) Unwrap() error {
	return e.Err
}

func GetErrorFromStatusFunc(service string) func(httpMethod, url, method string, status int, err error) error {
	return func(httpMethod, urlc, method string, status int, err error) error {
		if status >= 200 && status <= 299 {
			return nil
		}

		return ErrorClient{
			ServiceName: service,
			Path:        method,
			Url:         httpMethod + " " + urlc,
			Err:         err,
			Status:      status,
		}
	}
}

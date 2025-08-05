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

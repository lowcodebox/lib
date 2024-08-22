package app_lib

import (
	"context"
	"fmt"
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/api-client"
)

var config struct {

	// VFS
	VfsBucket      string `envconfig:"VFS_BUCKET" default:"buildbox"`
	VfsKind        string `envconfig:"VFS_KIND" default:"s3"`
	VfsEndpoint    string `envconfig:"VFS_ENDPOINT" default:"http://127.0.0.1:9000"`
	VfsAccessKeyId string `envconfig:"VFS_ACCESS_KEY_ID" default:"minioadmin"`
	VfsSecretKey   string `envconfig:"VFS_SECRET_KEY" default:"minioadmin"`
	VfsRegion      string `envconfig:"VFS_REGION" default:""`
	VfsComma       string `envconfig:"VFS_COMMA" default:""`
	VfsCertCA      string `envconfig:"VFS_CERT_CA" default:""`

	// LOGBOX
	LogboxEndpoint       string        `envconfig:"LOGBOX_ENDPOINT" default:"http://127.0.0.1:8999"`
	LogboxAccessKeyId    string        `envconfig:"LOGBOX_ACCESS_KEY_ID" default:""`
	LogboxSecretKey      string        `envconfig:"LOGBOX_SECRET_KEY" default:""`
	LogboxRequestTimeout time.Duration `envconnfig:"LOGBOX_REQUEST_TIMEOUT" default:"300ms"`

	// LOGBOX-client CircuitBreaker
	CbMaxRequestsLogbox uint32        `envconfig:"CB_MAX_REQUESTS_LOGBOX" default:"3" description:"максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии"`
	CbTimeoutLogbox     time.Duration `envconfig:"CB_TIMEOUT_LOGBOX" default:"5s" description:"период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние"`
	CbIntervalLogbox    time.Duration `envconfig:"CB_INTERVAL_LOGBOX" default:"5s" description:"циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков"`
}

func Test_funcMap_recurchildren(t1 *testing.T) {
	a := api.New(context.Background(), "https://lms.wb.ru/lms/api", false, time.Second, 100, time.Hour, time.Minute, "LKHlhb899Y09olUi")
	NewFuncMap(nil, a, nil, "", nil, nil, nil)

	result1 := Funcs.RecursiveChildren("2024-04-23T16-03-22z03-00-e09385", "leader", 0)
	fmt.Printf("%+v", result1)
}

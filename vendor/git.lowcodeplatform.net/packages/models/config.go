package models

import (
	"path/filepath"
	"strconv"
	"time"
)

const Sep = string(filepath.Separator)

// Float custom duration for toml configs
type Float struct {
	float64
	Value float64
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Float) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.ParseFloat(string(text), 10)
	d.Value = i
	return err
}

// Bool custom duration for toml configs
type Bool struct {
	bool
	Value bool
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Bool) UnmarshalText(text []byte) error {
	var err error
	d.Value = false
	if string(text) == "true" {
		d.Value = true
	}
	return err
}

// Duration custom duration for toml configs
type Duration struct {
	time.Duration
	Value time.Duration
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	t := string(text)
	// если получили только цифру - добавляем секунды (по-умолчанию)
	if len(t) != 0 {
		lastStr := t[len(t)-1:]
		if lastStr != "h" && lastStr != "m" && lastStr != "s" {
			t = t + "m"
		}
	}
	d.Value, err = time.ParseDuration(t)
	return err
}

// Int custom duration for toml configs
type Int struct {
	int
	Value int
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Int) UnmarshalText(text []byte) error {
	var err error
	tt := string(text)
	if tt == "" {
		d.Value = 0
		return nil
	}
	i, err := strconv.Atoi(tt)
	d.Value = i
	return err
}

// Config системный конфиг с общей структурой для всех сервисов
type Config struct {
	HttpsOnly  Bool   `envconfig:"HTTPS_ONLY" default:""`
	ProjectKey string `envconfig:"PROJECT_KEY" default:""`
	SignUrlKey string `envconfig:"SIGNIN_URL_KEY" default:""`

	Domain string `envconfig:"DOMAIN" default:""`
	Type   string `envconfig:"TYPE" default:"sender"`

	ServiceVersion string `envconfig:"SERVICE_VERSION" default:""`
	HashCommit     string `envconfig:"HASH_COMMIT" default:""`

	// Настройки размещение
	DC          string `envconfig:"DC" default:"msk"`
	Environment string `envconfig:"ENVIRONMENT" default:"dev"`
	Cluster     string `envconfig:"CLUSTER" default:"alpha"`
	ConfigID    string `envconfig:"CONFIG_ID" default:""`

	RunTime time.Time `envconfig:"RUN_TIME" default:""`
	UpTime  string    `envconfig:"UP_TIME" default:""`
	HashRun string    `envconfig:"HASH_RUN" default:"is empty"`

	// LOGBOX
	LogboxEndpoint      string   `envconfig:"LOGBOX_ENDPOINT" default:"127.0.0.1:8999"`
	CbMaxRequestsLogbox uint32   `envconfig:"CB_MAX_REQUESTS_LOGBOX" default:"3" description:"максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии"`
	CbTimeoutLogbox     Duration `envconfig:"CB_TIMEOUT_LOGBOX" default:"5s" description:"период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние"`
	CbIntervalLogbox    Duration `envconfig:"CB_INTERVAL_LOGBOX" default:"5s" description:"циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков"`

	// Http
	MaxRequestBodySize Int      `envconfig:"MAX_REQUEST_BODY_SIZE" default:"10485760"`
	ReadTimeout        Duration `envconnfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout       Duration `envconnfig:"WRITE_TIMEOUT" default:"10s"`
	ReadBufferSize     Int      `envconfig:"READ_BUFFER_SIZE" default:"16384"`

	Configuration string `envconfig:"CONFIGURATION" default:""`

	GRPC   Int `envconfig:"GRPC" default:"8998"`
	HTTP   Int `envconfig:"HTTP" default:"8080"`
	HTTPS  Int `envconfig:"HTTPS" default:"443"`
	MCP    Int `envconfig:"MCP" default:"8001"`
	Bridge Int `envconfig:"BRIDGE" default:"9000"`
}

// VFSConfig системный конфиг для подключения к VFS
type VFSConfig struct {
	VfsBucket      string `envconfig:"VFS_BUCKET" default:""`
	VfsKind        string `envconfig:"VFS_KIND" default:"s3"`
	VfsEndpoint    string `envconfig:"VFS_ENDPOINT" default:"http://127.0.0.1:9000"`
	VfsAccessKeyID string `envconfig:"VFS_ACCESS_KEY_ID" default:"minioadmin"`
	VfsSecretKey   string `envconfig:"VFS_SECRET_KEY" default:"minioadmin"`
	VfsRegion      string `envconfig:"VFS_REGION" default:""`
	VfsComma       string `envconfig:"VFS_COMMA" default:""`
	VfsCertCA      string `envconfig:"VFS_CERT_CA" default:"" description:"CA-сертификат"`
	VfsCAFile      string `envconfig:"VFS_CA_FILE" default:"" description:"Файл CA-сертификата"`
}

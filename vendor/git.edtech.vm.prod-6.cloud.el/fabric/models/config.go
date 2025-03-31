package models

import (
	"strconv"
	"strings"
	"time"
)

// Bool custom duration for toml configs
type Bool bool

// Float custom duration for toml configs
type Float float64

// Duration custom duration for toml configs
type Duration time.Duration

// Int custom duration for toml configs
type Int int

type PingConfig struct {
	Uid                 string `envconfig:"UID" default:""`
	Service             string `envconfig:"SERVICE" default:""`
	Projectuid          string `envconfig:"PROJECTUID" default:""`
	Project             string `envconfig:"PROJECT" default:"" description:"имя проекта"`
	ProjectPointsrc     string `envconfig:"PROJECT_POINTSRC" default:""`
	Name                string `envconfig:"NAME" default:"" description:"имя сервиса"`
	Version             string `envconfig:"VERSION" default:""`
	VersionPointsrc     string `envconfig:"VERSION_POINTSRC" default:""`
	HttpsOnly           Bool   `envconfig:"HTTPS_ONLY" default:"false"`
	UidService          string `envconfig:"UID_SERVICE" default:""`
	HashCommit          string `envconfig:"HASH_COMMIT" default:""`
	Environment         string `envconfig:"ENVIRONMENT" default:"dev"`
	EnvironmentPointsrc string `envconfig:"ENVIRONMENT_POINTSRC" default:"dev"`
	Cluster             string `envconfig:"CLUSTER" default:"alpha"`
	ClusterPointsrc     string `envconfig:"CLUSTER_POINTSRC" default:"alpha"`
	DC                  string `envconfig:"DC" default:"el"`
	AccessPublic        Bool   `envconfig:"ACCESS_PUBLIC" default:"false"`
	Port                Int    `envconfig:"PORT" default:"0"`
	PortHttp            Int    `envconfig:"PORT_HTTP" default:"0"`
	PortHttps           Int    `envconfig:"PORT_HTTPS" default:"0"`
	PortGrpc            Int    `envconfig:"PORT_GRPC" default:"0"`
	Replicas            Int    `envconfig:"REPLICAS" default:"0"`
	Follower            string `envconfig:"FOLLOWER" default:""`
	Mask                string `envconfig:"MASK" default:""`
}

type PingConfigOld struct {
	ServiceType            string `envconfig:"SERVICE_TYPE" default:"gui"`
	UidService             string `envconfig:"UID_SERVICE" default:""`
	ServiceVersion         string `envconfig:"SERVICE_VERSION" default:""`
	EnvironmentPointsrc    string `envconfig:"ENVIRONMENT_POINTSRC" default:"dev"`
	Cluster                string `envconfig:"CLUSTER" default:"alpha"`
	ClusterPointsrc        string `envconfig:"CLUSTER_POINTSRC" default:"alpha"`
	DataUid                string `envconfig:"DATA_UID" default:""`
	Domain                 string `envconfig:"DOMAIN" default:""`
	Port                   string `envconfig:"PORT" default:""`
	Projectuid             string `envconfig:"PROJECTUID" default:""`
	ProjectPointsrc        string `envconfig:"PROJECT_POINTSRC" default:""`
	VersionPointsrc        string `envconfig:"VERSION_POINTSRC" default:""`
	ReplicasService        Int    `envconfig:"REPLICAS_SERVICE" default:"0"`
	ServicePreloadPointsrc string `envconfig:"SERVICE_PRELOAD_POINTSRC" default:""`
}

// UnmarshalText method satisfying toml unmarshal interface
func (b *Bool) UnmarshalText(text []byte) error {
	*b = strings.ToLower(string(text)) == "true"

	return nil
}

func (b Bool) V() bool {
	return bool(b)
}

// UnmarshalText method satisfying toml unmarshal interface
func (f *Float) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.ParseFloat(string(text), 10)
	*f = Float(i)

	return err
}

func (f Float) V() float64 {
	return float64(f)
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Duration) UnmarshalText(text []byte) error {
	t := string(text)
	// если получили только цифру - добавляем секунды (по-умолчанию)
	if len(t) != 0 {
		lastStr := t[len(t)-1:]
		if lastStr != "h" && lastStr != "m" && lastStr != "s" {
			t = t + "m"
		}
	}

	parsed, err := time.ParseDuration(t)
	*d = Duration(parsed)

	return err
}

func (d Duration) V() time.Duration {
	return time.Duration(d)
}

// UnmarshalText method satisfying toml unmarshal interface
func (i *Int) UnmarshalText(text []byte) error {
	tt := string(text)
	if tt == "" {
		*i = 0

		return nil
	}

	v, err := strconv.Atoi(tt)
	*i = Int(v)

	return err
}

func (i Int) V() int {
	return int(i)
}

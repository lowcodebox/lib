package models

import "time"

// Bool custom duration for toml configs
type Bool struct {
	bool
	Value bool
}

type PingConfig struct {
	Uid                 string    `envconfig:"UID" default:""`
	Service             string    `envconfig:"SERVICE" default:""`
	Projectuid          string    `envconfig:"PROJECTUID" default:""`
	Project             string    `envconfig:"PROJECT" default:"" description:"имя проекта"`
	ProjectPointsrc     string    `envconfig:"PROJECT_POINTSRC" default:""`
	Name                string    `envconfig:"NAME" default:"" description:"имя сервиса"`
	Version             string    `envconfig:"VERSION" default:""`
	VersionPointsrc     string    `envconfig:"VERSION_POINTSRC" default:""`
	HttpsOnly           string    `envconfig:"HTTPS_ONLY" default:""`
	UidService          string    `envconfig:"UID_SERVICE" default:""`
	HashCommit          string    `envconfig:"HASH_COMMIT" default:""`
	Environment         string    `envconfig:"ENVIRONMENT" default:"dev"`
	EnvironmentPointsrc string    `envconfig:"ENVIRONMENT_POINTSRC" default:"dev"`
	RunTime             time.Time `envconfig:"RUN_TIME" default:""`
	UpTime              string    `envconfig:"UP_TIME" default:""`
	Cluster             string    `envconfig:"CLUSTER" default:"alpha"`
	ClusterPointsrc     string    `envconfig:"CLUSTER_POINTSRC" default:"alpha"`
	DC                  string    `envconfig:"DC" default:"el"`
	AccessPublic        Bool      `envconfig:"ACCESS_PUBLIC" default:"false"`
	Domain              string    `envconfig:"DOMAIN" default:""`
	Port                string    `envconfig:"PORT" default:""`
}

type PingConfigOld struct {
	Name                string    `envconfig:"NAME" default:"" description:"имя сервиса (по-умолчанию = тип сервиса)"`
	ServiceType         string    `envconfig:"SERVICE_TYPE" default:"gui"`
	Version             string    `envconfig:"VERSION" default:"app"`
	HttpsOnly           string    `envconfig:"HTTPS_ONLY" default:""`
	UidService          string    `envconfig:"UID_SERVICE" default:""`
	ServiceVersion      string    `envconfig:"SERVICE_VERSION" default:""`
	HashCommit          string    `envconfig:"HASH_COMMIT" default:""`
	Environment         string    `envconfig:"ENVIRONMENT" default:"dev"`
	EnvironmentPointsrc string    `envconfig:"ENVIRONMENT_POINTSRC" default:"dev"`
	RunTime             time.Time `envconfig:"RUN_TIME" default:""`
	UpTime              string    `envconfig:"UP_TIME" default:""`
	Cluster             string    `envconfig:"CLUSTER" default:"alpha"`
	ClusterPointsrc     string    `envconfig:"CLUSTER_POINTSRC" default:"alpha"`
	DC                  string    `envconfig:"DC" default:"el"`
	AccessPublic        Bool      `envconfig:"ACCESS_PUBLIC" default:"false"`
	DataUid             string    `envconfig:"DATA_UID" default:""`
	Domain              string    `envconfig:"DOMAIN" default:""`
	Port                string    `envconfig:"PORT" default:""`
	Projectuid          string    `envconfig:"PROJECTUID" default:""`
	ProjectPointsrc     string    `envconfig:"PROJECT_POINTSRC" default:""`
	VersionPointsrc     string    `envconfig:"VERSION_POINTSRC" default:""`
}

// UnmarshalText method satisfying toml unmarshal interface
func (b *Bool) UnmarshalText(text []byte) error {
	b.Value = false
	if string(text) == "true" {
		b.Value = true
	}

	return nil
}

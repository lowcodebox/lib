package session

import (
	"context"
	"sync"

	api "git.edtech.vm.prod-6.cloud.el/fabric/api-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	iam "git.edtech.vm.prod-6.cloud.el/fabric/iam-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

type session struct {
	ctx context.Context
	cfg model.Config
	api api.Api
	iam iam.IAM

	Registry SessionRegistry
}

type SessionRegistry struct {
	Mx *sync.RWMutex
	M  map[string]SessionRec
}

type SessionRec struct {
	UID      string             `json:"uid"`
	DeadTime int64              `json:"dead_time"`
	Profile  models.ProfileData `json:"profile"`
}

type Session interface {
	Found(sessionID string) (status bool)
	GetProfile(sessionID string) (profile *models.ProfileData, err error)
	Delete(sessionID string) (err error)
	Set(sessionID string) (err error)
	List() (result map[string]SessionRec)
	Cleaner(ctx context.Context) (err error)
}

func New(ctx context.Context, cfg model.Config, api api.Api, iam iam.IAM) Session {
	registrySession := SessionRegistry{
		Mx: &sync.RWMutex{},
	}

	return &session{
		ctx,
		cfg,
		api,
		iam,
		registrySession,
	}
}

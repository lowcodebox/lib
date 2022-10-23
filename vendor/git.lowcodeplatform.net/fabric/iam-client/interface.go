package client

import (
	"git.lowcodeplatform.net/fabric/iam/pkg/i18n"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

type iam struct {
	url			string
	projectKey 	string
	logger 		lib.Log
	metric 		lib.ServiceMetric
	msg  		i18n.I18n
}

type IAM interface {
	Verify(tokenString string) (statue bool, body *models.Token, refreshToken string, err error)
	Refresh(token, profile string, expire bool) (result string, err error)
	ProfileGet(sessionID string) (result string, err error)
	ProfileList() (result string, err error)
}

func New(url, projectKey string, logger lib.Log, metric lib.ServiceMetric) IAM {
	if url[len(url)-1:] == "/" {
		url = url[:len(url)-1]
	}
	msg := i18n.New()
	return &iam{
		url,
		projectKey,
		logger,
		metric,
		msg,
	}
}
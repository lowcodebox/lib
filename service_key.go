package lib

import (
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

const (
	defaultTokenInterval = time.Minute * 5
)

type XServiceKeyBuilder struct {
	key           models.XServiceKey
	tokenInterval time.Duration
}

func NewServiceKey() *XServiceKeyBuilder {
	return &XServiceKeyBuilder{
		tokenInterval: defaultTokenInterval,
	}
}

func (b *XServiceKeyBuilder) WithDomain(domain string) *XServiceKeyBuilder {
	b.key.Domain = domain
	return b
}

func (b *XServiceKeyBuilder) WithClient(client string) *XServiceKeyBuilder {
	b.key.Client = client
	return b
}

func (b *XServiceKeyBuilder) WithCheckCert(checkCert bool) *XServiceKeyBuilder {
	b.key.CheckCert = checkCert
	return b
}

func (b *XServiceKeyBuilder) WithWhiteURI(whiteURI string) *XServiceKeyBuilder {
	b.key.WhiteURI = whiteURI
	return b
}

func (b *XServiceKeyBuilder) WithRole(role string) *XServiceKeyBuilder {
	b.key.Role = role
	return b
}

func (b *XServiceKeyBuilder) WithProfile(profile string) *XServiceKeyBuilder {
	b.key.Profile = profile
	return b
}

func (b *XServiceKeyBuilder) WithGroups(groups string) *XServiceKeyBuilder {
	b.key.Groups = groups
	return b
}

func (b *XServiceKeyBuilder) WithServiceUID(serviceUID string) *XServiceKeyBuilder {
	b.key.ServiceUID = serviceUID
	return b
}

func (b *XServiceKeyBuilder) WithRequestID(requestID string) *XServiceKeyBuilder {
	b.key.RequestID = requestID
	return b
}

func (b *XServiceKeyBuilder) WithInterval(interval time.Duration) *XServiceKeyBuilder {
	b.tokenInterval = interval
	return b
}

func (b *XServiceKeyBuilder) Build(projectKey []byte) (token string, err error) {
	b.key.Expired = time.Now().Add(b.tokenInterval).Unix()
	return encodeServiceKey(b.key, projectKey)
}

func DecodeServiceKey(xServiceKey string, projectKey []byte) (out models.XServiceKey, err error) {
	return decodeServiceKey(projectKey, xServiceKey)
}

func EncodeServiceKey(in models.XServiceKey, projectKey []byte) (token string, err error) {
	return encodeServiceKey(in, projectKey)
}

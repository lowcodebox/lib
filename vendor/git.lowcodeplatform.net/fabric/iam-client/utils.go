package client

import (
	"encoding/json"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"github.com/golang-jwt/jwt"
)

func (s *iam) Verify(tokenString string) (statue bool, body *models.Token, refreshToken string, err error) {
	var in models.Token
	var jtoken = map[string]string{}

	jsonToken, err := lib.Decrypt([]byte(s.projectKey), tokenString)
	if err != nil {
		return false, nil, refreshToken, err
	}

	err = json.Unmarshal([]byte(jsonToken), &jtoken)
	if err != nil {
		return false, nil, refreshToken, err
	}

	tokenAccess := jtoken["access"]
	refreshToken = jtoken["refresh"]

	token, err := jwt.ParseWithClaims(tokenAccess, &in, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.projectKey), nil
	})

	if !token.Valid {
		return false, nil, refreshToken, s.msg.TokenValidateFail.Error()
	}
	tbody := token.Claims.(*models.Token)

	return true, tbody, refreshToken, err
}
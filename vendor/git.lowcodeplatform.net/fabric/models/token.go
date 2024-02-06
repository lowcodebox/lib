package models

import "github.com/golang-jwt/jwt"

type Token struct {
	Uid        string
	Role       string
	Profile    string
	Groups     string
	Local      string
	Type       string
	Session    string
	SessionRev string // ревизия текущей сессии (если сессия обновляется (меняется профиль) - ID-сессии остается, но ревизия меняется
	jwt.StandardClaims
}

type Roles struct {
	Title string
	Uid   string
}

type XServiceKey struct {
	Domain  string
	Expired int64
}

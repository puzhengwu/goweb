package goweb

import (
//"net/http"
)

type QqOauth struct {
	appid        string
	appkey       string
	redirecturl  string
	authorizeurl string
}

func NewQqOauth(cfg map[string]string) *QqOauth {
	oauth := &QqOauth{
		cfg["appid"],
		cfg["appkey"],
		cfg["redirecturl"],
		cfg["authorizeurl"],
	}
	return oauth
}

func (q *QqOauth) GetAuthorizeUrl() string {
	return q.authorizeurl + "?response_type=code&client_id=" + q.appid + "&routes.go"
}

package models

import (
	"time"

	"github.com/labstack/gommon/color"
)

var (
	Done    = color.Green("[DONE]")
	Process = color.Blue("[....]")
	Fail    = color.Red("[FAIL]")
	Warn    = color.Yellow("[WARN]")
)

type Alive struct {
	Uid      string `json:"uid"`
	Pid      int    `json:"pid"`
	Project  string `json:"project"`
	Path     string `json:"path"`
	Replicas int    `json:"replicas"`

	HTTP   int `json:"HTTP"`
	Grpc   int `json:"Grpc"`
	HTTPS  int `json:"HTTPS"`
	Bridge int `json:"BRIDGE"`
	MCP    int `json:"MCP"`

	Env        string    `json:"env"`
	Cluster    string    `json:"cluster"`
	DC         string    `json:"dc"`
	Mask       string    `json:"mask"`
	Uptime     string    `json:"uptime"`
	Runtime    time.Time `json:"runtime"`
	Public     bool      `json:"public"`
	Version    string    `json:"version"`
	HashCommit string    `json:"hash_commit"`

	Status string `json:"status"`
	Error  string `json:"error"`

	OS   string `json:"os"`
	Arch string `json:"arch"`
}

package grpcbalancer

import (
	"errors"
	"fmt"
)

const RoundRobin BalancingPolicy = iota

var ErrEmptyVal = errors.New("empty value")

type BalancingPolicy int8

type grpcConfig struct {
	Balancing BalancingPolicy `json:"loadBalancingPolicy"`
}

func (bp BalancingPolicy) String() string {
	switch bp {
	case RoundRobin:
		return "round_robin"
	}

	return ""
}

func (bp BalancingPolicy) MarshalJSON() ([]byte, error) {
	val := bp.String()
	if val == "" {
		return nil, ErrEmptyVal
	}

	return []byte(fmt.Sprintf(`"%s"`, val)), nil
}

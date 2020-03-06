package service

import (
	"go.uber.org/atomic"
	"time"
)

// HandleInfo ...
type HandleInfo struct {
	ServiceName string
	Data        interface{}
	Callback    HandleCallback
}

// HandleCallback ...
type HandleCallback func(src interface{})

// Node ...
type Node interface {
	Start()
}

var dateKey = time.Date(2019, time.November, 11, 10, 20, 10, 300, time.Local)

type node struct {
	lock *atomic.Bool
}

func nodeInstance() *node {
	return &node{lock: atomic.NewBool(false)}
}

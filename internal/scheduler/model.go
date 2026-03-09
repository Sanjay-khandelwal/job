package scheduler

import (
	"sync"
	"time"
)

type data struct {
	Interval time.Duration
	Task     func()
	stop     chan struct{}
}

type info map[string]*data

type Scheduler struct {
	jobs []*info
	mu   sync.Mutex
	wg   sync.WaitGroup
}

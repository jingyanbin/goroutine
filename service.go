package goroutine

import (
	"github.com/jingyanbin/basal"
	"github.com/jingyanbin/log"
	"sync"
	"sync/atomic"
)

type iService interface {
	Start()
	Stop()
	Started() bool
}

type services struct {
	services map[iService]bool
	mu       sync.Mutex
	gow      GoWaitGroup
	stopped  int32
}

func (my *services) Add(service iService) {
	my.mu.Lock()
	defer my.mu.Unlock()
	if _, ok := my.services[service]; ok {
		return
	}
	if service.Started() {
		my.services[service] = true
	} else {
		my.services[service] = false
	}
}

func (my *services) Start(service iService) bool {
	my.mu.Lock()
	defer my.mu.Unlock()
	if _, ok := my.services[service]; ok {
		return false
	}
	if service.Started() {
		my.services[service] = true
		return false
	}
	my.services[service] = true
	my.gow.Go(service.Start)
	log.InfoF("services started: %v", service)
	return true
}

func (my *services) StartAll() int {
	my.mu.Lock()
	defer my.mu.Unlock()
	var count int
	var size = len(my.services)
	for service, state := range my.services {
		if state == false {
			my.services[service] = true
			my.gow.Go(service.Start)
			count += 1
		}
	}
	log.InfoF("services started count: %v/%v", count, size)
	return count
}

func (my *services) Stop() {
	if !atomic.CompareAndSwapInt32(&my.stopped, 0, 1) {
		return
	}
	defer basal.Exception(func(stack string, e error) {
		log.ErrorF("services stop error: %s", stack)
	})
	my.mu.Lock()
	defer my.mu.Unlock()
	for service := range my.services {
		if service.Started() {
			service.Stop()
		}
	}
	my.gow.Wait()
	log.InfoF("services stopped count: %v", len(my.services))
}

func (my *services) Listening(stopping func() bool, blocking bool) {
	if blocking {
		StopSignal.Listening(stopping)
	} else {
		go StopSignal.Listening(stopping)
	}
}

var Services = &services{services: make(map[iService]bool)}

package goroutine

import (
	"github.com/jingyanbin/log"
	"sync"
)

type iService interface {
	Run()
	Stop()
}

type services struct {
	services map[iService]bool
	mu       sync.Mutex
	gow      GoWaitGroup
}

func (my *services) Register(service iService) {
	my.mu.Lock()
	defer my.mu.Unlock()
	if _, ok := my.services[service]; ok {
		return
	}
	my.services[service] = false
}

func (my *services) Start(service iService) bool {
	my.mu.Lock()
	defer my.mu.Unlock()
	if _, ok := my.services[service]; ok {
		return false
	}
	my.services[service] = true
	my.gow.Go(service.Run)
	log.InfoF("services started: %v", service)
	return true
}

func (my *services) Run() int {
	my.mu.Lock()
	defer my.mu.Unlock()
	var count int
	var size = len(my.services)
	for service, state := range my.services {
		if state == false {
			my.services[service] = true
			my.gow.Go(service.Run)
			count += 1
		}
	}
	log.InfoF("services running count: %v/%v", count, size)
	return count
}

func (my *services) Stop() {
	my.mu.Lock()
	defer my.mu.Unlock()
	var count int
	var size = len(my.services)
	for service, state := range my.services {
		if state == true {
			my.services[service] = false
			service.Stop()
			count += 1
		}
	}
	my.gow.Wait()
	log.InfoF("services stopped count: %v/%v", count, size)
}

func (my *services) StopSignalListening() {
	StopSignalListening(my.Stop)
}

var Services services

func init() {
	Services.services = map[iService]bool{}
}

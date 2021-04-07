package goroutine

import (
	. "github.com/jingyanbin/basal"
	"github.com/jingyanbin/log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type GoState struct {
	stopping chan struct{}
	closed   int32
}

func (m *GoState) close() {
	if atomic.CompareAndSwapInt32(&m.closed, 0, 1) {
		close(m.stopping)
	}
}

func (m *GoState) Running() bool {
	return atomic.LoadInt32(&m.closed) == 0
}

func (m *GoState) Stopping() <-chan struct{} {
	return m.stopping
}

func NewGoState() *GoState {
	return &GoState{stopping: make(chan struct{}, 0)}
}

type goroutines struct {
	counter int32
	wg      sync.WaitGroup
	state   *GoState
}

func (m *goroutines) add() {
	m.wg.Add(1)
	atomic.AddInt32(&m.counter, 1)
}

func (m *goroutines) done() {
	atomic.AddInt32(&m.counter, -1)
	m.wg.Done()
}

func (m *goroutines) close() {
	m.state.close()
}

func (m *goroutines) count() int32 {
	return atomic.LoadInt32(&m.counter)
}

func (m *goroutines) newGo(f func(), df func()) {
	m.add()
	go func() {
		defer m.done()
		if df != nil {
			defer df()
		}
		defer Exception(func(stack string, e error) {
			log.FatalF(stack)
		})
		f()
	}()
}

func (m *goroutines) newGoWith(f func(*GoState), df func(), gSta *GoState) {
	m.add()
	go func() {
		defer m.done()
		if df != nil {
			defer df()
		}
		defer Exception(func(stack string, e error) {
			log.FatalF(stack)
		})
		if gSta == nil {
			f(m.state)
		} else {
			f(gSta)
		}
	}()
}

func (m *goroutines) wait() {
	m.waitF(nil, 1)
}

func (m *goroutines) waitF(f func(), second time.Duration) {
	var waiting int32 = 1
	var wg sync.WaitGroup
	wg.Add(1)
	var sleepTime time.Duration
	if second > 0 {
		sleepTime = time.Second * second
	} else {
		sleepTime = time.Second
	}
	go func() {
		defer wg.Done()
		var cur int32
		for atomic.LoadInt32(&waiting) == 1 {
			count := atomic.LoadInt32(&m.counter)
			if cur != count {
				cur = count
				log.InfoF("goroutines waiting: %v/%v", count, runtime.NumGoroutine())
			}
			time.Sleep(sleepTime)
			if f != nil {
				f()
			}
		}
	}()
	m.wg.Wait()
	atomic.StoreInt32(&waiting, 0)
	wg.Wait()
	log.InfoF("goroutines stopped: %v/%v", m.count(), runtime.NumGoroutine())
}

var goMgr = &goroutines{state: NewGoState()}

func Go(f func()) {
	goMgr.newGo(f, nil)
}

func GoWith(f func(goState *GoState)) {
	goMgr.newGoWith(f, nil, nil)
}

func Count() int32 {
	return goMgr.count()
}

func Close() {
	goMgr.close()
}

func Wait() {
	goMgr.wait()
}

func WaitF(f func(), second time.Duration) {
	goMgr.waitF(f, second)
}

type GoWaitGroup struct {
	wg sync.WaitGroup
}

func (m *GoWaitGroup) add() {
	m.wg.Add(1)
}

func (m *GoWaitGroup) done() {
	m.wg.Done()
}

func (m *GoWaitGroup) Go(f func()) {
	m.add()
	goMgr.newGo(f, m.done)
}

func (m *GoWaitGroup) Wait() {
	m.wg.Wait()
}

type WaitClose struct {
	stopping    chan struct{}
	initialized int32
	closed      int32
}

func (m *WaitClose) init() {
	if atomic.CompareAndSwapInt32(&m.initialized, 0, 1) {
		m.stopping = make(chan struct{}, 0)
	}
}

func (m *WaitClose) Close() {
	m.init()
	if atomic.CompareAndSwapInt32(&m.closed, 0, 1) {
		close(m.stopping)
	}
}

func (m *WaitClose) Wait() {
	m.init()
	<-m.stopping
}

const runningStandby = 0
const runningStarted = 1
const runningStopped = 2

type RunningState struct {
	running int32
	wc      WaitClose
}

func (m *RunningState) Starting() bool {
	if atomic.CompareAndSwapInt32(&m.running, runningStandby, runningStarted) {
		return true
	} else {
		return false
	}
}

func (m *RunningState) Stopping() bool {
	if atomic.CompareAndSwapInt32(&m.running, runningStarted, runningStopped) {
		return true
	} else {
		return false
	}
}

func (m *RunningState) Close() {
	m.wc.Close()
}

func (m *RunningState) Wait() {
	m.wc.Wait()
}

//异常立即退出时调用
func Exit() {
	log.Wait()
	os.Exit(0)
}

//执行完成正常退出时调用
func Finalize() {
	goMgr.close()
	goMgr.wait()
	log.Wait()
}

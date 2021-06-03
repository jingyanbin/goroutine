package goroutine

import (
	"github.com/jingyanbin/basal"
	"github.com/jingyanbin/log"
	"os"
	"os/signal"
	"syscall"
)

type stopSignal struct {
}

func (m *stopSignal) Listening(stopping func() bool) {
	defer basal.Exception(func(stack string, e error) {
		log.ErrorF("listen stop signal error: %s", stack)
	})
	log.InfoF("listen stop signal start...")
	mySignal := make(chan os.Signal, 1)
	signal.Notify(mySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	for {
		select {
		case sig := <-mySignal:
			switch sig.(syscall.Signal) {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				log.InfoF("listen stopping signal: %v", sig)
				if stopping() {
					return
				}
			default:
				log.InfoF("listen other signal %v", sig)
			}
		}
	}
}

func (*stopSignal) Wait() {
	log.InfoF("wait stop signal start...")
	mySignal := make(chan os.Signal, 1)
	signal.Notify(mySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	for {
		select {
		case sig := <-mySignal:
			switch sig.(syscall.Signal) {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				log.InfoF("wait stopping signal: %v", sig)
				return
			default:
				log.InfoF("wait other signal: %v", sig)
			}
		case <-goMgr.state.stopping:
			log.InfoF("gowith goroutines state closed")
			return
		}
	}
}

func (m *stopSignal) ServicesWaitAfter() {

}

var StopSignal = &stopSignal{}

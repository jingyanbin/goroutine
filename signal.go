package goroutine

import (
	"github.com/jingyanbin/log"
	"os"
	"os/signal"
	"syscall"
)

func StopSignalListening(stopping func()) {
	mySignal := make(chan os.Signal, 1)
	signal.Notify(mySignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	for {
		select {
		case sig := <-mySignal:
			switch sig.(syscall.Signal) {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				log.InfoF("stopping signal %v", sig)
				stopping()
				log.InfoF("stopped signal %v", sig)
				return
			}
		}
	}
}
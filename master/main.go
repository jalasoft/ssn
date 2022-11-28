package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jalasoft/ssn/master/agentserver"
	"github.com/jalasoft/ssn/master/restserver"
)

var agentPort int = 10001
var restPort int = 10002

var interruptChan chan interface{}
var finalInterruptChan chan interface{}

var errorChan chan error
var infoChan chan string

func init() {
	interruptChan = make(chan interface{})
	finalInterruptChan = make(chan interface{})
	errorChan = make(chan error)
	infoChan = make(chan string)
}

func main() {

	flag.IntVar(&agentPort, "agent-port", 10001, "--agent-port 10001")
	flag.IntVar(&restPort, "rest-port", 10002, "--rest-port 10002")
	flag.Parse()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)

	go func() {
		<-signalChan
		close(interruptChan)
	}()

	go func() {
		for {
			select {
			case errorMsg := <-errorChan:
				log.Printf("ERROR: %v", errorMsg)
			case <-finalInterruptChan:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case infoMsg := <-infoChan:
				log.Printf("INFO: %s", infoMsg)
			case <-finalInterruptChan:
				return
			}
		}
	}()

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go agentserver.Listen(agentPort, infoChan, errorChan, interruptChan, &waitGroup)
	go restserver.Listen(restPort, infoChan, errorChan, interruptChan, &waitGroup)

	waitGroup.Wait()
	infoChan <- "Application is closing"
	close(finalInterruptChan)
}

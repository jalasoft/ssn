package agentserver

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/jalasoft/ssn/master/agentserver/protocol"
)

func Listen(port int, infoChan chan<- string, errorChan chan<- error, interruptChan <-chan interface{}, waitGroup *sync.WaitGroup) {
	infoChan <- fmt.Sprintf("Starting agent server listening on port %d", port)

	listener, err := net.ListenTCP("tcp4", &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	})

	if err != nil {
		errorChan <- fmt.Errorf("Could not start server on port %d: %v", port, err)
		return
	}

	infoChan <- "Awating connections of agents."

root:
	for {
		select {
		case <-interruptChan:
			infoChan <- "Stopping listenning for buddies."
			break root

		default:
			listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
			con, err := listener.AcceptTCP()

			if err != nil {
				if opError, ok := err.(*net.OpError); ok && opError.Timeout() {
					break
				}

				errorChan <- fmt.Errorf("Could not connect with client: %v", err)
			} else {
				session := protocol.NewSession(con, infoChan, errorChan)
				infoChan <- fmt.Sprintf("New session %s initiated.", session.Id())
				go protocol.HelloState(session, interruptChan)
			}
		}
	}
	infoChan <- "Buddy server stopped."
	waitGroup.Done()
}

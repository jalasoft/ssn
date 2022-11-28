package restserver

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jalasoft/ssn/master/registry"
)

var shutdownErr error = nil
var shutdownMutex sync.Mutex

func Listen(port int, infochanChan chan<- string, errorChan chan<- error, interruptChan chan interface{}, waitGroup *sync.WaitGroup) {
	infochanChan <- fmt.Sprintf("Starting REST server listening on port %d.", port)

	router := gin.Default()
	router.GET("/agent", allAgentsHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			shutdownMutex.Lock()
			shutdownErr = fmt.Errorf("Cannot start server on port %d: %v", port, err)
			shutdownMutex.Unlock()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-interruptChan:
		shutdownMutex.Lock()
		if shutdownErr == nil {
			if err := srv.Shutdown(ctx); err != nil {
				errorChan <- fmt.Errorf("Could not shutdown REST server: %v", err)
			}
		}
		shutdownMutex.Unlock()
		infochanChan <- "REST server stopped."
		waitGroup.Done()
	}

}

func allAgentsHandler(ctx *gin.Context) {

	registry.All()
	ctx.String(200, "Cajk")
}

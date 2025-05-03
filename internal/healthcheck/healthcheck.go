package healthcheck

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func RunHealthCheckFunction() {
	// start health check goroutine
	go func() {
		ticker := time.NewTicker(time.Minute / 2)
		defer ticker.Stop()

		client := &http.Client{}
		apiURL := os.Getenv("STAY_ALIVE_API_URL")
		if apiURL == "" {
			return
		}

		for range ticker.C {
			_, err := client.Get(fmt.Sprintf("%s/healthcheck", apiURL))
			if err != nil {
				log.Printf("health check failed: %v", err)
			} else {
				log.Printf("health check passed")
			}
		}
	}()
}

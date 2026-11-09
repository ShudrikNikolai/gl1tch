// Package healthcheck реализует self-check для distroless-образов.
package healthcheck

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// Maybe выполняет GET /healthz, если процесс запущен с аргументом healthcheck.
func Maybe(args []string, port int) {
	if len(args) < 2 || args[1] != "healthcheck" {
		return
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", port))
	if err != nil {
		os.Exit(1)
	}

	status := resp.StatusCode
	_ = resp.Body.Close()
	if status != http.StatusOK {
		os.Exit(1)
	}

	os.Exit(0)
}

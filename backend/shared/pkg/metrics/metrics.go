// Package metrics отдаёт минимальный Prometheus-текст без внешних зависимостей.
package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

var started = time.Now()

// Handler возвращает /metrics с info-меткой сервиса и базовой статистикой Go.
func Handler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = fmt.Fprintf(w, "# HELP gl1tch_service_info Build info.\n")
		_, _ = fmt.Fprintf(w, "# TYPE gl1tch_service_info gauge\n")
		_, _ = fmt.Fprintf(w, "gl1tch_service_info{service=%q} 1\n", service)
		_, _ = fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines.\n")
		_, _ = fmt.Fprintf(w, "# TYPE go_goroutines gauge\ngo_goroutines %d\n", runtime.NumGoroutine())
		_, _ = fmt.Fprintf(w, "# HELP go_memstats_alloc_bytes Allocated bytes.\n")
		_, _ = fmt.Fprintf(w, "# TYPE go_memstats_alloc_bytes gauge\ngo_memstats_alloc_bytes %d\n", m.Alloc)
		_, _ = fmt.Fprintf(w, "# HELP go_memstats_sys_bytes Total bytes from OS.\n")
		_, _ = fmt.Fprintf(w, "# TYPE go_memstats_sys_bytes gauge\ngo_memstats_sys_bytes %d\n", m.Sys)
		_, _ = fmt.Fprintf(w, "# HELP process_start_time_seconds Start time.\n")
		_, _ = fmt.Fprintf(w, "# TYPE process_start_time_seconds gauge\nprocess_start_time_seconds %d\n", started.Unix())
	}
}

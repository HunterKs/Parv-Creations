package middleware

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const logDir = "logs"

type dailyLogWriter struct {
	mu       sync.Mutex
	file     *os.File
	fileDate string
}

func (w *dailyLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateLocked(time.Now()); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *dailyLogWriter) rotateLocked(now time.Time) error {
	activeDate := now.Format("2006-01-02")
	if w.file != nil && w.fileDate == activeDate {
		return nil
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	nextFile, err := os.OpenFile(filepath.Join(logDir, activeDate+".log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if w.file != nil {
		_ = w.file.Close()
	}

	w.file = nextFile
	w.fileDate = activeDate
	return nil
}

// InitDailyLogger routes standard logs to stdout and the active date-stamped log file.
func InitDailyLogger() error {
	writer := &dailyLogWriter{}
	if err := writer.rotateLocked(time.Now()); err != nil {
		return err
	}

	log.SetFlags(0)
	log.SetOutput(io.MultiWriter(os.Stdout, writer))
	return nil
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

// LoggerMiddleware logs every HTTP request that reaches the router.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)

		log.Printf(
			"[%s] [%s] [%s] [%s] [%d]",
			time.Now().Format(time.RFC3339),
			r.Method,
			r.URL.String(),
			remoteIP(r.RemoteAddr),
			recorder.status,
		)
	})
}

func remoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

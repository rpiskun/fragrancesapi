package main

import (
	// "bytes"

	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	// "github.com/urfave/negroni"
)

// LoggerEntry is the structure
// passed to the template.
type LoggerEntry struct {
	StartTime string
	Status    int
	Duration  string
	Hostname  string
	Method    string
	Path      string
}

// LoggerDefaultFormat is the format
// logged used by the default Logger instance.
var LoggerDefaultFormat = "{{.StartTime}} | {{.Status}} | %8dms | {{.Method}}\t| {{.Path}}\n"

// LoggerDefaultDateFormat is the
// format used for date by the
// default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// ALogger interface
type ALogger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	// ALogger implements just enough log.Logger interface to be compatible with other implementations
	ALogger
	dateFormat     string
	template       *template.Template
	durationFormat string
}

// NewLogger returns a new Logger instance
func NewLogger() *Logger {
	logger := &Logger{ALogger: log.New(os.Stdout, "[api] ", 0), dateFormat: LoggerDefaultDateFormat}
	logger.SetFormat(LoggerDefaultFormat)
	return logger
}

func (l *Logger) SetFormat(format string) {
	l.template = template.Must(template.New("negroni_parser").Parse(format))
}

func (l *Logger) SetDateFormat(format string) {
	l.dateFormat = format
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// start := time.Now()

	next(rw, r)

	// res := rw.(negroni.ResponseWriter)
	// d := time.Since(start)
	// log := LoggerEntry{
	// 	StartTime: start.Format(l.dateFormat),
	// 	Status:    res.Status(),
	// 	Duration:  d.String(),
	// 	Hostname:  r.Host,
	// 	Method:    r.Method,
	// 	Path:      r.URL.Path,
	// }

	// buff := &bytes.Buffer{}
	// l.template.Execute(buff, log)
	// l.Printf(buff.String(), d.Nanoseconds()/int64(time.Millisecond))
	// l.Printf(
	// 	" %23s | %3d | %8dms | %6s | %s", 
	// 	log.StartTime,
	// 	log.Status, 
	// 	d.Nanoseconds()/int64(time.Millisecond),
	// 	log.Method,
	// 	log.Path)
	// l.Println(log.Status)
	// l.Printf(buff.String(), d.String())
}

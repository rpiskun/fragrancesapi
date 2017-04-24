package main

import (
	"log"
	"path/filepath"
	"runtime"
)

func TracePrint(s string) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Printf("[error] in %s[%s:%d] %s", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line, s)
}

func TracePrintError(err error) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Printf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line, err)
}

func TraceFatal(s string) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Fatalf("[fatal] in %s[%s:%d] %s", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line, s)
}

func TraceFatalError(err error) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Fatalf("[fatal] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line, err)
}

func TraceDebug(s string) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Printf("[debug] in %s[%s:%d] %s", runtime.FuncForPC(pc).Name(), filepath.Base(fn), line, s)
}

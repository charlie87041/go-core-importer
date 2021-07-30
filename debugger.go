package main

import (
	"fmt"
	"time"
)

type Debugger interface {
	debug(message ...interface{})
	error(message ...interface{})
	getErrors() []interface{}
	getStatus() int
	chrono(text string)
	setTimer(timer time.Time)
}

type defaultdebugger struct {
	Debugger
	isDebugEnabled bool
	timer          time.Time
	status         int
	errors         []interface{}
}

func (ptr *defaultdebugger) debug(message ...interface{}) {
	if ptr.isDebugEnabled {
		fmt.Println(message...)
	}
}

func (ptr *defaultdebugger) setTimer(time time.Time) {
	ptr.timer = time
}

func (ptr *defaultdebugger) getErrors() []interface{} {
	return ptr.errors
}

func (ptr *defaultdebugger) getStatus() int {
	return ptr.status
}
func (ptr *defaultdebugger) error(message ...interface{}) {
	if ptr.isDebugEnabled {
		fmt.Println(message...)
	}
	ptr.status = 1
	ptr.errors = append(ptr.errors, message...)

}

func (ptr *defaultdebugger) chrono(msg string) {
	if ptr.isDebugEnabled {
		fmt.Println(msg, time.Since(ptr.timer))
	}
}

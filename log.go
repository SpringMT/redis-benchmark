package main

import (
	"fmt"
	"log"
	"strings"
)

type loglevel int

const (
	mute loglevel = iota
	warn
	info
	debug
)

func (rb *redisBench) logLevel() loglevel {
	return loglevel(len(rb.Verbose))
}

func (rb *redisBench) logf(lv loglevel, format string, a ...interface{}) {
	rb.log(lv, fmt.Sprintf(format, a...))
}

func (rb *redisBench) log(lv loglevel, str string) {
	if rb.logLevel() < lv || lv <= mute {
		return
	}
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	log.Print(str)
}

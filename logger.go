package main

import (
	"fmt"
	"time"
)

// Logger - logs things as they develop
type Logger struct {
	start   time.Time
	lastLog time.Time
	silent  bool
}

func (l *Logger) log(str string) {
	now := time.Now()
	duration := now.Sub(l.start)
	if !l.silent {
		fmt.Printf("\n[ Time elapsed: %s]\n", duration.String())
		fmt.Println(str)
	}
	l.lastLog = now
}

// ConstructLogger - constructor
func ConstructLogger(mode string) Logger {
	return Logger{time.Now(), time.Now(), mode == "silent"}
}

package main

import (
	"fmt"
	"time"
)

// StringLL - linked list for strings
type StringLL struct {
	crt  string
	next *StringLL
}

func (ll *StringLL) add(str string) *StringLL {
	ll.crt = str
	ll.next = &StringLL{}
	return ll.next
}

// Logger - logs things as they progress
type Logger struct {
	start     time.Time
	lastLog   time.Time
	mode      string
	writes    *StringLL
	lastWrite *StringLL
}

// Log - logs a string
func (l *Logger) Log(str string) {
	if l.mode == "silent" {
		return
	}
	now := time.Now()
	duration := now.Sub(l.start)
	s := fmt.Sprintf("\n[ Time elapsed: %s]\n%s\n", duration.String(), str)
	if l.mode == "print" {
		fmt.Printf(s)
	} else if l.mode == "write" {
		l.lastWrite = l.lastWrite.add(s)
	}
	l.lastLog = now
}

// LogAll - log everything from a list of strings
func (l *Logger) LogAll(header string, items []string) {
	s := header
	for _, item := range items {
		s += "\n\t" + item
	}
	l.Log(s)
}

// ConstructLogger - constructor
func ConstructLogger(mode string) *Logger {
	writeStart := &StringLL{} // pointer to beginning of LL
	return &Logger{time.Now(), time.Now(), mode, writeStart, writeStart}
}

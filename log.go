package main

import (
	"fmt"
	"os"
)

type Logger struct {
	path string
	file *os.File
}

func NewLogger(path string) *Logger {
	var file, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND, os.ModeAppend|os.ModePerm)
	if err != nil {
		panic(err)
	}
	var l = new(Logger)
	l.path = path
	l.file = file
	return l
}

func (l *Logger) Write(s string) {
	l.file.WriteString(s + "\n")
}

func (l *Logger) Writef(s string, a ...interface{}) {
	var content = fmt.Sprintf(s, a...)
	l.file.WriteString(content + "\n")
}

func (l *Logger) Close() {
	l.file.Close()
}

package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

var (
	Info  *log.Logger
	Error *log.Logger
)

type Fields map[string]any

func Init() {
	Info = log.New(os.Stdout, "", log.LstdFlags)
	Error = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
}

func InfoJSON(message string, fields Fields) {
	writeJSON("info", message, fields)
}

func ErrorJSON(message string, fields Fields) {
	writeJSON("error", message, fields)
}

func writeJSON(level string, message string, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}

	event := Fields{
		"level":     level,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	}

	for key, value := range fields {
		event[key] = value
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf(`{"level":"error","message":"failed to marshal log event","error":%q}`, err.Error())
		return
	}

	if level == "error" {
		Error.Println(string(data))
		return
	}

	Info.Println(string(data))
}

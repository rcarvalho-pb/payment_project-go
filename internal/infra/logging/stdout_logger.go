package logging

import (
	"encoding/json"
	"fmt"
	"time"
)

type StdoutLogger struct {
	Stream *Stream
}

func NewStdoutLogger(stream *Stream) StdoutLogger {
	return StdoutLogger{Stream: stream}
}

func (l *StdoutLogger) log(level, msg string, fields map[string]any) {
	clonedFields := cloneFields(fields)
	if l.Stream != nil {
		l.Stream.Publish(Entry{
			Level:   level,
			Message: msg,
			Time:    time.Now().UTC().Format(time.RFC3339),
			Fields:  clonedFields,
		})
	}

	entry := map[string]any{
		"level": level,
		"msg":   msg,
		"time":  time.Now().UTC().Format(time.RFC3339),
	}

	for k, v := range clonedFields {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	fmt.Println(string(b))
}

func (l *StdoutLogger) Info(msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

func (l *StdoutLogger) Error(msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}

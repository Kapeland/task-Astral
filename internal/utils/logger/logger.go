package logger

import (
	"fmt"
	"time"
)

const logTimeLayout = "2006-01-02 15:04:05.00"

const (
	InfoPrefix  = "INFO"
	DebugPrefix = "DEBUG"
	WarnPrefix  = "WARN"
	ErrPrefix   = "ERROR"
)

func Log(prefix string, message string) {
	out := createLogMessage(time.Now(), prefix, message)
	fmt.Println(out)
}

func createLogMessage(time1 time.Time, prefix string, message string) string {
	return fmt.Sprintf("%s\t%s\t%s", time1.Format(logTimeLayout), prefix, message)
}

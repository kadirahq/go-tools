package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	goerr "github.com/go-errors/errors"
)

const (
	delim = ","
	envar = "log"
)

var (
	levels = map[string]bool{}
	logAll = false
	logger = New("app")
	output = log.New(os.Stdout, "", log.LstdFlags)
	colblu = color.New(color.FgBlue).SprintFunc()
	colred = color.New(color.FgRed).SprintFunc()
	colyel = color.New(color.FgYellow).SprintFunc()
	colcya = color.New(color.FgCyan).SprintFunc()
)

func init() {
	env := os.Getenv(envar)
	if env == "" {
		env = "info,error,time"
	}

	for _, lvl := range strings.Split(env, delim) {
		levels[lvl] = true
	}
}

// Enable enables a level
func Enable(lvl string) {
	levels[lvl] = true
}

// Disable disables a level
func Disable(lvl string) {
	levels[lvl] = false
}

// Print prints any level logs using the default logger
func Print(lvl string, logs ...interface{}) {
	logger.Print(lvl, logs...)
}

// Info prints info level logs using the default logger
func Info(logs ...interface{}) {
	logger.Info(logs...)
}

// Debug prints debug level logs using the default logger
func Debug(logs ...interface{}) {
	logger.Debug(logs...)
}

// Error prints error logs using the default logger
func Error(err error, logs ...interface{}) {
	logger.Error(err, logs...)
}

// Time tracks the time duration using the default logger
func Time(beg time.Time, logs ...interface{}) {
	logger.Time(beg, logs...)
}

// Logger is a logger with a header
type Logger struct {
	head string
}

// New creates a new logger
func New(head string) *Logger {
	return &Logger{head}
}

// New creates a logger extending the header
func (l *Logger) New(head string) *Logger {
	return &Logger{l.head + ":" + head}
}

// Print prints log items if given level is enabled
func (l *Logger) Print(lvl string, logs ...interface{}) {
	if levels[lvl] {
		content := fmt.Sprintf("("+lvl+") "+l.head+": %+v", logs)
		output.Println(content)
	}
}

// Info prints basic information to stdout
func (l *Logger) Info(logs ...interface{}) {
	if levels["info"] {
		content := fmt.Sprintf("%s: %+v", colblu("(info) "+l.head), logs)
		output.Println(content)
	}
}

// Debug logs lots of details useful for debugging
func (l *Logger) Debug(logs ...interface{}) {
	if levels["error"] {
		content := fmt.Sprintf("%s: %+v", colyel("(debug) "+l.head), logs)
		output.Println(content)
	}
}

// Error prints an error and some additional information
func (l *Logger) Error(err error, logs ...interface{}) {
	if levels["error"] {
		content := fmt.Sprintf("%s: %+v", colred("(error) "+l.head), logs)

		switch e := err.(type) {
		case *goerr.Error:
			output.Println(content + "\n" + e.ErrorStack())
		default:
			output.Println(content + "\n" + err.Error())
		}
	}
}

// Time tracks the time duration from start time
// This is best when used with a defer statement.
func (l *Logger) Time(beg time.Time, logs ...interface{}) {
	if levels["time"] {
		dur := time.Since(beg)
		content := fmt.Sprintf("%s: %s %+v", colcya("(time) "+l.head), dur, logs)
		output.Println(content)
	}
}

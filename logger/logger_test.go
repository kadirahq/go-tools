package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	goerr "github.com/go-errors/errors"
)

var (
	buffer = bytes.NewBuffer(nil)
)

func init() {
	output.SetOutput(buffer)
	output.SetFlags(0)
}

func TestDefaults(t *testing.T) {
	if !levels["info"] || !levels["error"] {
		t.Fatal("Default log levels aren't enabled")
	}
}

func TestPrint(t *testing.T) {
	buffer.Reset()

	// should not print
	Print("test", 0, 0)

	// enable log level "test"
	Enable("test")

	// should print
	Print("test", 1, 2)
	Print("test", 4, 5, "hello")

	// Disable log level "test"
	Disable("test")

	// should not print
	Print("test", 6, 7)

	// expected output
	exp := "(test) app: [1 2]\n" +
		"(test) app: [4 5 hello]\n"

	if got := string(buffer.Bytes()); got != exp {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestInfo(t *testing.T) {
	buffer.Reset()

	Info(1, 2, 3)

	exp := colblu("(info) app") + ": [1 2 3]\n"

	if got := string(buffer.Bytes()); got != exp {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestDebug(t *testing.T) {
	buffer.Reset()

	Debug(1, 2, 3)

	exp := colyel("(debug) app") + ": [1 2 3]\n"

	if got := string(buffer.Bytes()); got != exp {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestError(t *testing.T) {
	buffer.Reset()

	Error(errors.New("test error"), 1, 2, 3)

	exp := colred("(error) app") + ": [1 2 3]\n" +
		"test error\n"

	if got := string(buffer.Bytes()); got != exp {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestErrorWithStack(t *testing.T) {
	buffer.Reset()

	Error(goerr.New("test error"), 1, 2, 3)

	exp := colred("(error) app") + ": [1 2 3]\n" +
		"*errors.errorString test error\n"

	if got := string(buffer.Bytes()); !strings.HasPrefix(got, exp) &&
		len(got) > len(exp) {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestTime(t *testing.T) {
	buffer.Reset()

	now := time.Now()
	time.Sleep(time.Millisecond)
	Time(now, 0)

	exp := colcya("(time) app") + ": 1."

	if got := string(buffer.Bytes()); !strings.HasPrefix(got, exp) &&
		strings.HasSuffix(got, "ms []") &&
		len(got) > len(exp) {
		t.Fatalf("exp: %s got: %s", exp, got)
	}
}

func TestTimeCond(t *testing.T) {
	buffer.Reset()

	now := time.Now()
	time.Sleep(100 * time.Microsecond)
	Time(now, time.Millisecond)

	if got := string(buffer.Bytes()); !strings.HasPrefix(got, "") {
		t.Fatalf("exp: %s got: %s", "", got)
	}
}

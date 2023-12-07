package lg

import (
	"strings"
	"testing"
)

func _fmt(lvl, msg string) string {
	return lvl + " " + msg
}

func replacePrint() (buf *strings.Builder, restore func()) {
	buf = &strings.Builder{}
	rawPrint := printFunc
	printFunc = func(lvl, message string) {
		buf.WriteString(_fmt(lvl, message))
	}
	return buf, func() {
		printFunc = rawPrint
	}
}

func TestSetVerbose(t *testing.T) {
	buf := &strings.Builder{}
	rawPrint := printFunc
	printFunc = func(_, message string) {
		buf.WriteString(message)
	}
	defer func() {
		printFunc = rawPrint
	}()
	SetVerbose(true)
	Debug("debugmsg")
	if buf.String() != "debugmsg" {
		t.Errorf("expect defaultPrint \"debugmsg\" if set verbose to true, but got:\n%q", buf.String())
	}
}

func TestInfo(t *testing.T) {
	buf, restore := replacePrint()
	defer restore()
	Info("msg")
	actual := buf.String()
	expect := _fmt("INFO", "msg")
	if actual != expect {
		t.Errorf("expect %q, got %q", expect, actual)
	}
}

func TestWarn(t *testing.T) {
	buf, restore := replacePrint()
	defer restore()
	Warn("msg")
	actual := buf.String()
	expect := _fmt("WARN", "msg")
	if actual != expect {
		t.Errorf("expect %q, got %q", expect, actual)
	}
}

func TestError(t *testing.T) {
	buf, restore := replacePrint()
	defer restore()
	Error("msg")
	actual := buf.String()
	expect := _fmt("ERROR", "msg")
	if actual != expect {
		t.Errorf("expect %q, got %q", expect, actual)
	}
}

package console

import (
	"os"

	"fmt"

	"golang.org/x/crypto/ssh/terminal"

	"bufio"
	"syscall"
	"strings"
)

var Trace = false

var reader = bufio.NewReader(os.Stdin)

func Write(msg ...interface{}) {
	fmt.Fprint(os.Stderr, msg...)
}

func Writeln(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
}
func Traceln(msg ...interface{}) {
	if (Trace) {
		fmt.Fprintln(os.Stderr, msg...)
	}
}

func ReadLine(msg string) string {
	fmt.Fprintf(os.Stderr, msg);

	text, _, _ := reader.ReadLine();
	return strings.TrimSpace(string(text));

}
func ReadDefault(msg string, defval string) string {
	fmt.Fprintf(os.Stderr, msg + " [" + defval + "]: ");
	b, _, _ := reader.ReadLine();
    text := string(b)
	if (text == "") {
		return defval
	}

	return text;

}
func Password(msg string) string {
	fmt.Fprintf(os.Stderr, msg);
	pw, _ := terminal.ReadPassword(syscall.Stdin)
	return string(pw)
}

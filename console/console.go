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

var NoMask = false


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
	return strings.TrimSpace(string(text))

}
func ReadDefault(msg string, defval string) string {
	fmt.Fprintf(os.Stderr, msg + " [" + defval + "]: ");
	b, _, _ := reader.ReadLine()
    text := string(b)
	if (text == "") {
		return defval
	}

	return text;

}
func Password(msg string) string {
	if (NoMask) {
	    return ReadLine(msg)

    } else {
        fmt.Fprintf(os.Stderr, msg);
        pw, err := terminal.ReadPassword(int(syscall.Stdin))
        if (err != nil) {
            // terminal may not be available, so input password in clear text
            pw, _, _ = reader.ReadLine();
        }
        // terminal password input eats newline,
        fmt.Fprintln(os.Stderr)
        return string(pw)
    }
}

// Craig Tomkow
// August 15, 2019

package configuration

import (
	"os"
	"testing"
)

var argTests = []struct {
	input		string
	expected	bool
}{
	{"install", true},
	{"remove", true},
	{"start", true},
	{"stop", true},
	{"status", true},
}

func TestCommand_MakeCmd(t *testing.T) {

	cmd := new(Command)

	for _, argTest := range argTests {
		os.Args = []string{"tto", argTest.input}
		err := cmd.MakeCmd()

		switch argTest.input {
		case "install":
			if argTest.expected != cmd.Install {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Install, argTest.expected)
			}
		case "remove":
			if argTest.expected != cmd.Remove {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Remove, argTest.expected)
			}
		case "start":
			if argTest.expected != cmd.Start {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Start, argTest.expected)
			}
		case "stop":
			if argTest.expected != cmd.Stop {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Stop, argTest.expected)
			}
		case "status":
			if argTest.expected != cmd.Status {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Status, argTest.expected)
			}
		default:
			if err == nil {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Status, argTest.expected)
			}
		}

	}
}

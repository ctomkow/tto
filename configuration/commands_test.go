// Craig Tomkow
// August 15, 2019

package configuration

import (
	"flag"
	"testing"
)

var testArgs = []struct {
	input		[]string
	expected	bool
}{
	{[]string{"install"}, true},
	{[]string{"remove"}, true},
	{[]string{"start"}, true},
	{[]string{"stop"}, true},
	{[]string{"status"}, true},
	{[]string{"fg"}, true},
	{[]string{"derp"}, false},
	{[]string{"dum", "dum"}, false},
	{[]string{""}, false},
}

func TestCommand_MakeCmd(t *testing.T) {

	cmd := new(Command)

	for _, argTest := range testArgs {

		_ = flag.CommandLine.Parse(argTest.input)
		err := cmd.MakeCmd()

		switch argTest.input[0] {
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
		case "fg":
			if argTest.expected != cmd.Fg {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Fg, argTest.expected)
			}
		case "derp":
			if err == nil {
				t.Errorf("Input arg test failed; found, expected: %#v, %s", err, "nil err")
			}
		case "dum":
			if err == nil {
				t.Errorf("Input arg test failed; found, expected: %#v, %s", err, "nil err")
			}
		case "":
			if err == nil {
				t.Errorf("Input arg test failed; found, expected: %#v, %s", err, "nil err")
			}
		default:
			if err == nil {
				t.Errorf("Input arg test failed; found, expected: %t, %t", cmd.Status, argTest.expected)
			}
		}
	}
}

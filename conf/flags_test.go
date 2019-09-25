// Craig Tomkow
// August 16, 2019

package conf

import (
	"flag"
	//"os"
	"testing"
)

func Test_ParseFlags(t *testing.T) {

	ParseFlags()
}

func Test_SetConfFlag(t *testing.T) {

	SetConfFlag()

	flagPtr := flag.Lookup("conf")
	if flagPtr == nil {
		t.Errorf("Set conf flag failed; found, expected: %#v, %s", flagPtr, "not nil ptr")
	}
}

// TODO: test is currently broken. It doesn't test failure scenarios at all. I thought flag.Usage sends to stderr...
func Test_SetUserUsage(t *testing.T) {

	SetUserUsage("usage", "commands", "flags")
	flag.Usage()

	/*

		var errBuff []byte
		numOfBytes, _ := os.Stderr.Read(errBuff)

		if numOfBytes == 0 {
			t.Errorf("Usage test failed; found, expected: %d, %s", numOfBytes, "not zero")
		}

	*/
}

func Test_SetLogToStderr(t *testing.T) {

	if err := SetLogToStderr(); err != nil {
		t.Errorf("Set log to stderr test failed; found, expected: %#v, %s", err, "nil err")
	}
}

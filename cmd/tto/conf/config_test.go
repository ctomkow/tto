// Craig Tomkow
// September 19, 2019

package conf

import (
	"net"
	"testing"
)

func TestConfig_MakeConfig(t *testing.T) {

	conf := new(Config)
	conf.MakeConfig()

	// top level config tests
	if !(conf.System.User == "username") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.User, "username")
	}
	if !(conf.System.Pass == "password") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.User, "password")
	}
	if !(conf.System.WorkingDir == "/opt/tto/") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.User, "/opt/tto/")
	}
	if !(conf.System.Type == "sender|receiver") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.User, "sender|receiver")
	}

	// sender config tests
	if !(conf.System.Role.Sender.Dest.IP.Equal(net.IP{6, 6, 6, 6})) {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.Dest.IP.String(), net.IP{6, 6, 6, 6}.String())
	}
	if !(conf.System.Role.Sender.Port == 22) {
		t.Errorf("Make config test failed; found, expected: %d, %d", conf.System.Role.Sender.Port, 22)
	}
	if !(conf.System.Role.Sender.Database == "mysql") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.Database, "mysql")
	}
	if !(conf.System.Role.Sender.DBip.IP.Equal(net.IP{7, 7, 7, 7})) {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.DBip.IP.String(), net.IP{7, 7, 7, 7}.String())
	}
	if !(conf.System.Role.Sender.DBport == 3306) {
		t.Errorf("Make config test failed; found, expected: %d, %d", conf.System.Role.Sender.Port, 3306)
	}
	if !(conf.System.Role.Sender.DBuser == "username") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.DBuser, "username")
	}
	if !(conf.System.Role.Sender.DBpass == "password") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.DBpass, "password")
	}
	if !(conf.System.Role.Sender.DBname == "databaseName") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.DBname, "databaseName")
	}
	if !(conf.System.Role.Sender.Cron == "a cron statement") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Sender.Cron, "a cron statement")
	}
	if !(conf.System.Role.Sender.MaxBackups == 5) {
		t.Errorf("Make config test failed; found, expected: %d, %d", conf.System.Role.Sender.MaxBackups, 5)
	}

	// receiver config tests
	if !(conf.System.Role.Receiver.Database == "mysql") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Receiver.Database, "mysql")
	}
	if !(conf.System.Role.Receiver.DBip.IP.Equal(net.IP{8, 8, 8, 8})) {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Receiver.DBip.IP.String(), net.IP{8, 8, 8, 8}.String())
	}
	if !(conf.System.Role.Receiver.DBport == 3306) {
		t.Errorf("Make config test failed; found, expected: %d, %d", conf.System.Role.Receiver.DBport, 3306)
	}
	if !(conf.System.Role.Receiver.DBuser == "username") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Receiver.DBuser, "username")
	}
	if !(conf.System.Role.Receiver.DBpass == "password") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Receiver.DBpass, "password")
	}
	if !(conf.System.Role.Receiver.DBname == "databaseName") {
		t.Errorf("Make config test failed; found, expected: %s, %s", conf.System.Role.Receiver.DBname, "databaseName")
	}
	if len(conf.System.Role.Receiver.ExecBefore) == 0 {
		t.Errorf("Make config test failed; found, expected: %d, %d", len(conf.System.Role.Receiver.ExecBefore), 0)
	}
	if len(conf.System.Role.Receiver.ExecAfter) == 0 {
		t.Errorf("Make config test failed; found, expected: %d, %d", len(conf.System.Role.Receiver.ExecAfter), 0)
	}
}

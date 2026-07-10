//go:build linux

package system

import "testing"

func TestLinuxNotifySendTerminatesOptionsBeforeTitle(t *testing.T) {
	opts := defaultNotificationOptions()
	opts.title = "--icon=attacker-controlled"
	opts.body = "body"

	command, err := linuxNotifySendCommand(opts, "")
	if err != nil {
		t.Fatalf("build notify-send command: %v", err)
	}
	if len(command.args) < 3 {
		t.Fatalf("unexpected notify-send arguments: %#v", command.args)
	}
	separator := -1
	for i, arg := range command.args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator < 0 || separator+1 >= len(command.args) || command.args[separator+1] != opts.title {
		t.Fatalf("expected option terminator immediately before title, got %#v", command.args)
	}
}

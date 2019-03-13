package v1alpha1

import (
	"testing"
)

func TestCstorVolumeCommand(t *testing.T) {
	tests := map[string]struct {
		action            RunCommandAction
		isSupportedAction bool
	}{
		"test 101": {DeleteCommandAction, false},
		"test 102": {CreateCommandAction, false},
		"test 103": {ListCommandAction, false},
		"test 104": {GetCommandAction, false},
		"test 105": {PatchCommandAction, false},
		"test 106": {UpdateCommandAction, false},
		"test 107": {ResizeCommandAction, true},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := WithAction(Command(), mock.action)
			c := &cstorVolumeCommand{cmd}
			result := c.Run()

			if !mock.isSupportedAction && result.Error() != ErrorNotSupportedAction {
				t.Fatalf("Test '%s' failed: expected 'ErrorNotSupportedAction': actual '%s': result '%s'", name, result.Error(), result)
			}
			if mock.isSupportedAction && result.Error() == ErrorNotSupportedAction {
				t.Fatalf("Test '%s' failed: expected 'supported action': actual 'ErrorNotSupportedAction': result '%s'", name, result)
			}
		})
	}
}

func TestvalidateOptions(t *testing.T) {
	tests := map[string]struct {
		cstorVolCmd    *cstorVolumeCommand
		isValidCommand bool
	}{
		"Empty volume name": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("127.0.1"), "volname": RunCommandData(""), "capacity": RunCommandData("10G")},
				},
			},
			isValidCommand: false,
		},
		"Empty IP": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData(""), "volname": RunCommandData("vol1"), "capacity": RunCommandData("20G")},
				},
			},
			isValidCommand: false,
		},
		"Empty Capacity": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("127.0.1"), "volname": RunCommandData("vol1"), "capacity": RunCommandData("")},
				},
			},
			isValidCommand: false,
		},
		"Populate all the values": {
			cstorVolCmd: &cstorVolumeCommand{
				RunCommand: &RunCommand{
					Data: RunCommandDataMap{"ip": RunCommandData("0.0.0.0"), "volname": RunCommandData("vol1"), "capacity": RunCommandData("5Zi")},
				},
			},
			isValidCommand: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := test.cstorVolCmd.validateOptions()
			if !test.isValidCommand && err == nil {
				t.Errorf("Expected error in command but got '%v'", err)
			}
		})
	}
}

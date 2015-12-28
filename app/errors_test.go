package app

import (
	"fmt"
	"testing"
)

func TestBuilderErr(t *testing.T) {
	err := builderErr(Null, fmt.Errorf("error foo bar"))
	if err.Error() != "null builder error: error foo bar" {
		t.Errorf("Expected \"null builder error: error foo bar\", got %q", err)
	}
}

func TestCommandFileErr(t *testing.T) {
	err := commandFileErr("test_command", "test/file.command", fmt.Errorf("error foo bar"))
	if err.Error() != "extracting commands for test_command from test/file.command failed: error foo bar" {
		t.Errorf("Expected \"extracting commands for test_command from test/file.command failed: error foo bar\", got %q", err)
	}
}

func TestDependentSettingErr(t *testing.T) {
	err := dependentSettingErr("foo", "bar")
	if err.Error() != "setting foo found but setting bar was not found-both are required" {
		t.Errorf("Expected \"setting foo found but setting bar was not found-both are required\", got %q", err)
	}
}

func TestMergeCommonSettingsErr(t *testing.T) {
	err := mergeCommonSettingsErr(fmt.Errorf("error foo bar"))
	if err.Error() != "merge of common settings failed: error foo bar" {
		t.Errorf("Expected \"merge of common settings failed: error foo bar\", got %q", err)
	}
}

func TestMergeSettingsErr(t *testing.T) {
	err := mergeSettingsErr(fmt.Errorf("error foo bar"))
	if err.Error() != "merge of section settings failed: error foo bar" {
		t.Errorf("Expected \"merge of section settings failed: error foo bar\", got %q", err)
	}
}

func TestNoCommandsFoundErr(t *testing.T) {
	err := noCommandsFoundErr("test_setting", "test/file.command")
	if err.Error() != "no commands for test_setting were found in test/file.command" {
		t.Errorf("Expected \"no commands for test_setting were found in test/file.command\", got %q", err)
	}
}

func TestProvisionerErr(t *testing.T) {
	err := provisionerErr(ShellScript, fmt.Errorf("error foo bar"))
	if err.Error() != fmt.Sprintf("%s provisioner error: error foo bar", ShellScript.String()) {
		t.Errorf("expected \"%s provisioner error: error foo bar\" got %q", ShellScript.String(), err)
	}
}

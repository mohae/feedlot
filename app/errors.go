package app

import (
	"fmt"
)

func builderErr(b Builder, err error) error {
	return fmt.Errorf("%s builder error: %s", b.String(), err.Error())
}

func commandFileErr(s, path string, err error) error {
	return fmt.Errorf("extracting commands for %s from %s failed: %s", s, path, err.Error())
}

func configNotFoundErr() error {
	return fmt.Errorf("configuration not found")
}

func dependentSettingErr(s1, s2 string) error {
	return fmt.Errorf("setting %s found but setting %s was not found-both are required", s1, s2)
}

func mergeCommonSettingsErr(err error) error {
	return fmt.Errorf("merge of common settings failed: %s", err.Error())
}

func mergeSettingsErr(err error) error {
	return fmt.Errorf("merge of section settings failed: %s", err.Error())
}

func noCommandsFoundErr(s, path string) error {
	return fmt.Errorf("no commands for %s were found in %s", s, path)
}

func provisionerErr(p Provisioner, err error) error {
	return fmt.Errorf("%s provisioner error: %s", p.String(), err.Error())
}

func requiredSettingErr(s string) error {
	return fmt.Errorf("required setting not found: %s", s)
}

func settingErr(s string, err error) error {
	return fmt.Errorf("encountered a problem processing the %s setting: %s", s, err.Error())
}

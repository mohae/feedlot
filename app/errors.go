package app

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrEmptyParam        = errors.New("received an empty paramater, expected a value")
)

// archivePriorBuildErr is a helper function to help generate consistent
// errors
func archivePriorBuildErr(err error) error {
	return fmt.Errorf("archive of prior build failed: %s", err)
}

func builderErr(b Builder, err error) error {
	return fmt.Errorf("%s builder error: %s", b.String(), err)
}

func commandFileErr(s, path string, err error) error {
	return fmt.Errorf("extracting commands for %s from %s failed: %s", s, path, err)
}

func configNotFoundErr() error {
	return fmt.Errorf("configuration not found")
}

func decodeErr(name string, err error) error {
	return fmt.Errorf("decode of %q failed: %s", name, err)
}

func dependentSettingErr(s1, s2 string) error {
	return fmt.Errorf("setting %s found but setting %s was not found-both are required", s1, s2)
}

func filenameNotSetErr(target string) error {
	return fmt.Errorf("%q not set, unable to retrieve the %s file", target, target)
}

func mergeCommonSettingsErr(err error) error {
	return fmt.Errorf("merge of common settings failed: %s", err)
}

func mergeSettingsErr(err error) error {
	return fmt.Errorf("merge of section settings failed: %s", err)
}

func noCommandsFoundErr(s, path string) error {
	return fmt.Errorf("no commands for %s were found in %s", s, path)
}

func provisionerErr(p Provisioner, err error) error {
	return fmt.Errorf("%s provisioner error: %s", p.String(), err)
}

func postProcessorErr(p PostProcessor, err error) error {
	return fmt.Errorf("%s post-processor error: %s", p.String(), err)
}


func PackerCreateErr(name string, err error) error {
	return fmt.Errorf("create of Packer template for %q failed: %s", name, err)
}

type RancherError struct {
	BuildName string
	Distro    string
	Operation string
	Problem   string
}

func (e RancherError) Error() string {
	return fmt.Sprintf("%s: %s %s, %s", e.BuildName, e.Distro, e.Operation, e.Problem)
}

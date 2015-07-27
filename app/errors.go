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

func requiredSettingErr(s string) error {
	return fmt.Errorf("required setting not found: %s", s)
}

func settingErr(s string, err error) error {
	return fmt.Errorf("encountered a problem processing the %s setting: %s", s, err)
}

func PackerCreateErr(name string, err error) error {
	return fmt.Errorf("create of Packer template for %q failed: %s", name, err)
}

func emptyPageErr(name, operation string) error {
	return ReleaseError{Name: name, Operation: operation, Problem: "page was empty"}
}

func checksumNotFoundErr(name, operation string) error {
	return ReleaseError{Name: name, Operation: operation, Problem: "checksum not found on page"}
}

func checksumNotSetErr(name string) error {
	return ReleaseError{Name: name, Operation: "setISOChecksum", Problem: "checksum not set"}
}

func noArchErr(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "arch was not set"}
}

func noFullVersionErr(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "full version was not set"}
}

func noMajorVersionErr(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "major version was not set"}
}

func noMinorVersionErr(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "minor version was not set"}
}

func noReleaseErr(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "release was not set"}
}

func setVersionInfoErr(name string, err error) error {
	return ReleaseError{Name: name, Operation: "SetVersionInfo", Problem: err.Error()}
}

func unsupportedReleaseErr(d Distro, name string) error {
	return fmt.Errorf("%s %s: unsupported release", d, name)
}

func osTypeBuilderErr(name, typ string) error {
	return ReleaseError{Name: name, Operation: "getOSType", Problem: fmt.Sprintf("%s is not supported by this distro", typ)}
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

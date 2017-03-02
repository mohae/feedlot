package app

import (
	"testing"

	"github.com/mohae/contour"
)

func TestBuildPackerTemplateFromDistros(t *testing.T) {
	_, err := buildPackerTemplateFromDistro()
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "get template: unsupported distro: " {
			t.Errorf("Expected \"get template: unsupported distro: \", got %q", err)
		}
	}
	contour.UpdateString("distro", "slackware")
	_, err = buildPackerTemplateFromDistro()
	if err.Error() != "get template: unsupported distro: slackware" {
		t.Errorf("Expected \"get template: unsupported distro: slackware\", got %q", err)
	}
}

func TestBuildPackerTemplateFromNamedBuild(t *testing.T) {
	doneCh := make(chan error)
	go buildPackerTemplateFromNamedBuild("", doneCh)
	err := <-doneCh
	if err == nil {
		t.Error("Expected an error, received none")
	} else {
		if err.Error() != "unable to build Packer template: no build name was received" {
			t.Errorf("Expected \"unable to build Packer template: no build name was received\", got %q", err)
		}
	}
	contour.RegisterString("build", "../test_files/conf/builds_test.toml")
	go buildPackerTemplateFromNamedBuild("", doneCh)
	err = <-doneCh
	if err == nil {
		t.Error("Expected an error, received none")
	} else {
		if err.Error() != "unable to build Packer template: no build name was received" {
			t.Errorf("Expected \"unable to build Packer template: no build name was received\", got %q", err)
		}
	}
	close(doneCh)
}

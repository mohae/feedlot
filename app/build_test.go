package app

import (
	"testing"

	"github.com/mohae/contour"
)

func TestBuildPackerTemplateFromDistros(t *testing.T) {
	a := ArgsFilter{}
	err := buildPackerTemplateFromDistro(a)
	err = buildPackerTemplateFromDistro(a)
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "unable to build Packer template: distro wasn't specified" {
			t.Errorf("Expected \"unable to build Packer template: distro wasn't specified\", got %q", err)
		}
	}
	a.Distro = "slackware"
	err = buildPackerTemplateFromDistro(a)
	if err.Error() != "unsupported distro: slackware" {
		t.Errorf("Expected \"unsupported distro: invalid\", got %q", err)
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
	contour.RegisterString(Build, "../test_files/conf/builds_test.toml")
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

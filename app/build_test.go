package app

import (
	"reflect"
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

func TestCopy(t *testing.T) {
	b := build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Type: "vagrant",
					Settings: []string{
						"output = :out_dir/packer.box",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"cleanup_test.sh",
						},
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
		},
	}

	bNew := b.copy()
	if !reflect.DeepEqual(b, bNew) {
		t.Errorf("expected the copies to be equal, they were not")
		return
	}
	// modify the copy, the original should not be affected
	bNew.BuilderIDs = append(bNew.BuilderIDs, "virtualbox-ovf")
	tsB := bNew.Builders["common"]
	tsB.templateSection.Settings = append(tsB.Settings, "boot_command = boot_test.command")
	bNew.Builders["common"] = tsB
	if len(b.Builders["common"].templateSection.Settings) == len(tsB.templateSection.Settings) {
		t.Error("the original builder was affected by the updates to new, it should not have been")
		return
	}

	bNew.PostProcessorIDs = append(bNew.PostProcessorIDs, "vagrant-cloud")
	tsPP := bNew.PostProcessors["vagrant"]
	tsPP.templateSection.Settings = append(tsPP.templateSection.Settings, "foo = bar")
	bNew.PostProcessors["vagrant"] = tsPP
	if len(b.PostProcessors["vagrant"].templateSection.Settings) == len(tsPP.templateSection.Settings) {
		t.Error("the original post-processor was affected by the updates to new, it should not have been")
		return
	}

	bNew.ProvisionerIDs = append(bNew.ProvisionerIDs, "shell-local")
	bNew.Provisioners["shell-local"] = bNew.Provisioners["shell"]
	if len(b.Provisioners) == len(bNew.Provisioners) {
		t.Error("the original provisioners was affected by the updates to new, it should not have been")
	}

}

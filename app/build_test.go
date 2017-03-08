package app

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

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
		if err.Error() != "build packer template failed: no build name was received" {
			t.Errorf("Expected \"build packer template failed: no build name was received\", got %q", err)
		}
	}
	contour.RegisterString("build", "../test_files/conf/builds_test.toml")
	go buildPackerTemplateFromNamedBuild("", doneCh)
	err = <-doneCh
	if err == nil {
		t.Error("Expected an error, received none")
	} else {
		if err.Error() != "build packer template failed: no build name was received" {
			t.Errorf("Expected \"build Packer template failed: no build name was received\", got %q", err)
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

	newBuild := b.Copy()
	msg, ok := EvalStringSlice(newBuild.BuilderIDs, b.BuilderIDs)
	if !ok {
		t.Errorf("BuilderIDs: %s", msg)
	}

	msg, ok = EvalBuilders(newBuild.Builders, b.Builders)
	if !ok {
		t.Errorf("Builders: %s", msg)
	}

	msg, ok = EvalStringSlice(newBuild.PostProcessorIDs, b.PostProcessorIDs)
	if !ok {
		t.Errorf("PostProcessorIDs: %s", msg)
	}

	msg, ok = EvalPostProcessors(newBuild.PostProcessors, b.PostProcessors)
	if !ok {
		t.Errorf("PostProcessors: %s", msg)
	}

	msg, ok = EvalStringSlice(newBuild.ProvisionerIDs, b.ProvisionerIDs)
	if !ok {
		t.Errorf("ProvisionerIDs: %s", msg)
	}

	msg, ok = EvalProvisioners(newBuild.Provisioners, b.Provisioners)
	if !ok {
		t.Errorf("Provisioners: %s", msg)
	}

}

func EvalBuilders(new, old map[string]builder) (msg string, ok bool) {
	if len(new) != len(old) {
		return fmt.Sprintf("copy length was %d; want %d", len(new), len(old)), false
	}
	for k, v := range old {
		vv, ok := new[k]
		if !ok {
			return fmt.Sprintf("expected copy to have entry for %q; none found", k), false
		}
		msg, ok := EvalTemplateSection(&vv.templateSection, &v.templateSection)
		if !ok {
			return fmt.Sprintf("%q: %s", k, msg), false
		}
	}
	return "", true
}

func EvalPostProcessors(new, old map[string]postProcessor) (msg string, ok bool) {
	if len(new) != len(old) {
		return fmt.Sprintf("copy length was %d; want %d", len(new), len(old)), false
	}
	for k, v := range old {
		vv, ok := new[k]
		if !ok {
			return fmt.Sprintf("expected copy to have entry for %q; none found", k), false
		}
		msg, ok := EvalTemplateSection(&vv.templateSection, &v.templateSection)
		if !ok {
			return fmt.Sprintf("%q: %s", k, msg), false
		}
	}
	return "", true
}

func EvalProvisioners(new, old map[string]provisioner) (msg string, ok bool) {
	if len(new) != len(old) {
		return fmt.Sprintf("copy length was %d; want %d", len(new), len(old)), false
	}
	for k, v := range old {
		vv, ok := new[k]
		if !ok {
			return fmt.Sprintf("expected copy to have entry for %q; none found", k), false
		}
		msg, ok := EvalTemplateSection(&vv.templateSection, &v.templateSection)
		if !ok {
			return fmt.Sprintf("%q: %s", k, msg), false
		}
	}
	return "", true
}

// This only checks Settings and Arrays. For arrays, only []string and
// [][]string, and map[string]string are supported.
func EvalTemplateSection(new, old *templateSection) (msg string, ok bool) {
	msg, ok = EvalStringSlice(new.Settings, old.Settings)
	if !ok {
		return msg, ok
	}
	for k, v := range old.Arrays {
		vv, ok := new.Arrays[k]
		if !ok {
			return fmt.Sprintf("Arrays %q: expected it to exist in the copy, it did not", k), false
		}
		switch v.(type) {
		case []string:
			msg, ok := EvalStringSlice(vv.([]string), v.([]string))
			if !ok {
				return fmt.Sprintf("Arrays %q: %s", k, msg), false
			}
		case [][]string:
			x := v.([][]string)
			xx := vv.([][]string)
			for i, val := range x {
				msg, ok := EvalStringSlice(xx[i], val)
				if !ok {
					return fmt.Sprintf("Arrays %q: index %v: %s", k, i, msg), false
				}
			}
		case map[string]string:
			x := v.(map[string]string)
			xx := vv.(map[string]string)
			for kk, val := range x {
				y, ok := xx[kk]
				if !ok {
					return fmt.Sprintf("Arrays %q: %q: expected it to exist in the copy, it did not", k, kk), false
				}
				if val != y {
					return fmt.Sprintf("Arrays: %q: %q: got %s; want %s", k, kk, y, val), false
				}
			}
		default:
			// anything not one of the above is out of scope and ignored.
		}

	}
	return "", true
}

func EvalStringSlice(new, old []string) (msg string, ok bool) {
	if (*reflect.SliceHeader)(unsafe.Pointer(&new)).Data == (*reflect.SliceHeader)(unsafe.Pointer(&old)).Data {
		return "expected slice data pointers to point to different locations; they didn't", false
	}
	if len(new) != len(old) {
		return fmt.Sprintf("expected slices to have the same length, got %d and %d", len(new), len(old)), false
	}
	for i, v := range new {
		if v != old[i] {
			return fmt.Sprintf("%d: got %v; want %v", i, v, old[i]), false
		}
	}
	return "", true
}

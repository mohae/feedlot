package app

import (
	"testing"
)

func TestDirWalk(t *testing.T) {
	tst := Archive{}
	err := tst.DirWalk("")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if tst.Files != nil {
			t.Error("Expected Archive.Files to be empty, got %q", tst.Files)
		}
	}

	err = tst.DirWalk("invalid/path/to/archive")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "archive of prior build failed: invalid/path/to/archive does not exist" {
			t.Errorf("Expected \"archive of prior build failed: invalid/path/to/archive does not exist\", got %q", err.Error())
		}
	}
}

func TestAddFilename(t *testing.T) {
	tst := Archive{}
	err := tst.addFilename("", "../test_files/src/ubuntu/scripts/dne_test_file.sh", nil, nil)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if "archive of prior build failed: ../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist" != err.Error() {
			t.Errorf("Expected \"archive of prior build failed: ../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist\", got %q", err.Error())
		}
	}
}

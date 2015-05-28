package app

import (
	"testing"
	"time"
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
		if err.Error() != "invalid/path/to/archive does not exist" {
			t.Errorf("Expected \"invalid/path/to/archive does not exist\", got %q", err.Error())
		}
	}
}

func TestAddFilename(t *testing.T) {
	tst := Archive{}
	err := tst.addFilename("", "../test_files/src/ubuntu/scripts/dne_test_file.sh", nil, nil)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if "../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist" != err.Error() {
			t.Errorf("Expected \"../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist\", got %q", err.Error())
		}
	}
}

func TestFormattedNow(t *testing.T) {
	fmtDateTime := formattedNow()
	// This doesn't feel right, as it just replicated the actual function
	// but I don't know how else to generate the needed dynamic value
	// and doing it this way will at least detect changes to the formula.
	dateTime := time.Now().Local().Format("2006-01-02T150405Z0700")
	if dateTime != fmtDateTime {
		t.Errorf("Expected %v, got %v", fmtDateTime, dateTime)
	}
}

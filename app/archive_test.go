package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mohae/contour"
)

func TestNewArchive(t *testing.T) {
	tests := []struct {
		name string
	}{
		{""},
		{"testArchive"},
	}
	for i, test := range tests {
		a := NewArchive(test.name)
		if a.Name != test.name {
			t.Errorf("%d: expected name to be %q got %q", i, test.name, a.Name)
		}
		if a.OutDir != "" {
			t.Errorf("%d: expected OutDir to be empty, was %q", i, a.OutDir)
		}
		if a.Type != "" {
			t.Errorf("%d: expected OutDir to be empty, was %q", i, a.Type)
		}
	}
}

func TestPriorBuild(t *testing.T) {
	// being in example mode should not cause errors
	// directory to be archived either missing or not a valid dir shouldn't cause errors
	tests := []struct {
		archivePriorBuild bool
		example           bool
		dir               string
		expectedFiles     []string
		expectedErr       error
	}{
		{false, false, "", []string{}, nil},
		{true, false, "", []string{}, nil},
		{true, false, "tst", []string{}, nil},                              // invvalid path
		{true, true, "tst", []string{}, nil},                               // invalid path
		{true, false, "test", []string{"test-0", "test-1", "test-2"}, nil}, // valid path
		{true, true, "test", []string{}, nil},                              // valid path
	}
	egOrig := contour.GetBool(Example)
	pbOrig := contour.GetBool(ArchivePriorBuild)
	for i, test := range tests {
		var dir string
		var files []string
		var err error
		if test.dir != "" && test.dir != "tst" {
			dir, files, err = createTmpTestDirFiles(fmt.Sprintf("feedlot-testpriorbuild-%d-", i))
			if err != nil {
				t.Errorf("error creating tmp files for test %d; aborting this test: %s", i, err)
				continue
			}
		}
		a := NewArchive(fmt.Sprintf("%s-%d", test.dir, i))
		contour.UpdateBool(Example, test.example)
		contour.UpdateBool(ArchivePriorBuild, test.archivePriorBuild)
		err = a.priorBuild(filepath.Join(dir, test.dir))
		if err != nil {
			if test.expectedErr == nil {
				t.Errorf("%d: expected no error, got %q", i, err)
				continue
			}
			if test.expectedErr.Error() != err.Error() {
				t.Errorf("%d expected error to be %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		tball := fmt.Sprintf("%s.tar.gz", filepath.Join(dir, a.Name))
		// if there aren't any expected files, then an archive shouldn't have been created
		if test.example || len(test.expectedFiles) == 0 {
			_, err = os.Stat(tball)
			if err == nil {
				t.Errorf("%d: expected an error; got none", i)
				continue
			}
			if !os.IsNotExist(err) {
				t.Errorf("%d: expected os.IsNotExist(); got %q", i, err)
			}
			goto deletetmp
		}
		if len(a.directory.Files) != len(files) {
			t.Errorf("%d: expected %d files to be indexed, %d were indexed.", i, len(test.expectedFiles), len(files))
			continue
		}
		for _, f := range a.directory.Files {
			if !stringSliceContains(files, filepath.Join(dir, "test", f.p)) {
				t.Errorf("%d: %s was indexed but the file is not found in the test.expectedFile slice", i, filepath.Join(dir, "test", f.p))
			}
		}
		_, err = os.Stat(tball)
		if err != nil {
			t.Errorf("%d: expected to find file %q; got %q", i, tball, err)
			continue
		}
	deletetmp:
		if dir != "" {
			err = os.RemoveAll(dir)
			if err != nil {
				t.Errorf("expected error to be nil; got %q", err)
			}
		}
	}
	contour.UpdateBool(Example, egOrig)
	contour.UpdateBool(ArchivePriorBuild, pbOrig)
}

func TestDirWalk(t *testing.T) {
	tst := Archive{}
	err := tst.DirWalk("")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if tst.Files != nil {
			t.Errorf("Expected Archive.Files to be empty, got %q", tst.Files)
		}
	}

	err = tst.DirWalk("invalid/path/to/archive")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "archive error: invalid/path/to/archive does not exist" {
			t.Errorf("Expected \"archive error: invalid/path/to/archive does not exist\", got %q", err)
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
			t.Errorf("Expected \"../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist\", got %q", err)
		}
	}
}

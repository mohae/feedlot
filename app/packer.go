package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	json "github.com/mohae/customjson"
	jww "github.com/spf13/jwalterweatherman"
)

type packerTemplate struct {
	Description      string                 `json:"description"`
	MinPackerVersion string                 `json:"min_packer_version"`
	Builders         []interface{}          `json:"builders,omitempty"`
	PostProcessors   []interface{}          `json:"post-processors,omitempty"`
	Provisioners     []interface{}          `json:"provisioners,omitempty"`
	Variables        map[string]interface{} `json:"variables,omitempty"`
}

// create a Packer build template based on the current configuration. The
// template is written to the output directory and any external resources that
// the template requires is copied there.
func (p *packerTemplate) create(i IODirInf, b BuildInf, dirs, files map[string]string) error {
	i.check()
	// priorBuild handles both the archiving and deletion of the prior build, if it exists, i.e.
	// if the build's output path exists.
	a := NewArchive(b.BuildName)
	err := a.priorBuild(appendSlash(i.OutputDir))
	if err != nil {
		jww.ERROR.Println(err)
		return PackerCreateErr(b.BuildName, err)
	}
	// create the destination directory if it doesn't already exist
	err = os.MkdirAll(i.OutputDir, 0754)
	if err != nil {
		return PackerCreateErr(b.BuildName, err)
	}
	// copy any directories associated with the template
	for dst, src := range dirs {
		fmt.Printf("CopyDir: %s to %s\n", src, dst)
		err = copyDir(src, dst)
		if err != nil {
			jww.ERROR.Println(err)
			return PackerCreateErr(b.BuildName, err)
		}
	}
	// copy the files associated with the template
	for dst, src := range files {
		fmt.Printf("CopyFiles: %s to %s\n", src, dst)
		_, err = copyFile(src, dst)
		if err != nil {
			jww.ERROR.Println(err)
			return PackerCreateErr(b.BuildName, err)
		}
	}
	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		jww.ERROR.Print(err)
		return PackerCreateErr(b.BuildName, err)
	}
	fname := filepath.Join(i.OutputDir, fmt.Sprintf("%s.json", b.Name))
	f, err := os.Create(fname)
	if err != nil {
		jww.ERROR.Print(err)
		return PackerCreateErr(b.BuildName, err)
	}
	// Close the file with error handling
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			jww.ERROR.Print(err)
			err = PackerCreateErr(b.BuildName, err)
		}
	}()
	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		jww.ERROR.Print(err)
		return PackerCreateErr(b.BuildName, err)
	}
	return nil
}

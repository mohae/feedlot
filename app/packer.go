package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mohae/feedlot/log"
	json "github.com/mohae/unsafejson"
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
func (p *packerTemplate) create(i IODirInf, b BuildInf, dirs, files map[string]string) (err error) {
	i.check()
	// priorBuild handles both the archiving and deletion of the prior build, if it exists, i.e.
	// if the build's output path exists.
	a := NewArchive(b.BuildName)
	err = a.priorBuild(appendSlash(i.TemplateOutputDir))
	if err != nil {
		err = Error{b.BuildName, err}
		log.Error(err)
		return err
	}
	// create the destination directory if it doesn't already exist
	err = os.MkdirAll(i.TemplateOutputDir, 0754)
	if err != nil {
		err = Error{b.BuildName, err}
		log.Error(err)
		return err
	}
	// copy any directories associated with the template
	for dst, src := range dirs {
		err = copyDir(src, dst)
		if err != nil {
			err = Error{b.BuildName, err}
			log.Error(err)
			return err
		}
	}
	// copy the files associated with the template
	for dst, src := range files {
		_, err = copyFile(src, dst)
		if err != nil {
			err = Error{b.BuildName, err}
			log.Error(err)
			return err
		}
	}
	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		err = Error{b.BuildName, err}
		log.Error(err)
		return err
	}
	fname := filepath.Join(i.TemplateOutputDir, fmt.Sprintf("%s.json", b.Name))
	f, err := os.Create(fname)
	if err != nil {
		err = Error{b.BuildName, err}
		log.Error(err)
		return err
	}
	// Close the file with error handling
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = Error{b.BuildName, cerr}
			log.Error(err)
		}
	}()

	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		err = Error{b.BuildName, err}
		log.Error(err)
		return err
	}
	return nil
}

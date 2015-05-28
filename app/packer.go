package app

import (
	"fmt"
	"io"
	"os"
	"sync"

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
	var wg sync.WaitGroup
	a := Archive{}
	wg.Add(1)
	err := a.priorBuild(appendSlash(i.OutDir), "gzip", &wg)
	wg.Wait()
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	// copy any directories associated with the template
	for dst, src := range dirs {
		err = copyDir(src, dst)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	// copy the files associated with the template
	for dst, src := range files {
		_, err = copyFile(src, dst)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	f, err := os.Create(appendSlash(i.OutDir) + fmt.Sprintf("%s.json", b.Name))
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	// Close the file with error handling
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			jww.ERROR.Print(err)
			err = cerr
		}
	}()
	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	return nil
}

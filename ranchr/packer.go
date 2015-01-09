package ranchr

import (
	_ "errors"
	"io"
	"os"
	"strconv"
	"time"

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
func (p *packerTemplate) create(i IODirInf, b BuildInf, scripts []string) error {
	err := i.check()
	if err != nil {
		jww.ERROR.Println("packerTemplate.create: " + err.Error())
		return err
	}
	// priorBuild handles both the archiving and deletion of the prior build, if it exists, i.e.
	// if the build's output path exists.
	a := Archive{}
	err = a.priorBuild(appendSlash(i.OutDir), "gzip")
	if err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}
	// TODO This needs to be handled better...this is too long for most builds but if there are situations
	// where there is a large archive this is not long enough.
	time.Sleep(time.Millisecond * 2000)
	err = copyFiles(scripts, i.ScriptsSrcDir, appendSlash(i.OutDir)+i.ScriptsDir)
	if err != nil {
		jww.ERROR.Println("packerTemplate.create: " + err.Error())
		return err
	}
	err = copyDirContent(i.HTTPSrcDir, appendSlash(i.OutDir)+i.HTTPDir)
	if err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}
	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}
	f, err := os.Create(appendSlash(i.OutDir) + b.Name + ".json")
	if err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}
	// Close the file with error handling
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			jww.ERROR.Print("packerTemplate.create: " + cerr.Error())
			err = cerr
		}
	}()
	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}
	return nil
}

func copyFiles(files []string, src string, dest string) error {
	var errCnt, okCnt int
	var err error
	for _, file := range files {
		_, err = copyFile(file, src, dest)
		if err != nil {
			jww.ERROR.Print(err)
			errCnt++
			continue
		}
		okCnt++
	}
	if errCnt > 0 {
		jww.ERROR.Print("copy of files for build had " + strconv.Itoa(errCnt) + " errors. There were " + strconv.Itoa(okCnt) + " files that were copied without error.")
		return err
	}
	return nil
}

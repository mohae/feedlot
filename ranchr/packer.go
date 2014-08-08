// packer.go contains the definition for packerTemplate along with its methods.
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

// A packerTemplate contains the final definition for a Packer build template.
// This is what is marshalled to JSON.
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
	jww.DEBUG.Println("packerTemplate.create: enter")
	jww.TRACE.Printf("%v/n%v/n%v", i, b, json.MarshalToString(scripts))

	if err := i.check(); err != nil {
		jww.ERROR.Println("packerTemplate.create: " + err.Error())
		return err
	}

	// priorBuild handles both the archiving and deletion of the prior build, if it exists, i.e.
	// if the build's output path exists.
	a := Archive{}

	if err := a.priorBuild(appendSlash(i.OutDir), "gzip"); err != nil {
		jww.ERROR.Print("packerTemplate.create: " + err.Error())
		return err
	}

	// TODO This needs to be handled better...this is too long for most builds but if there are situations
	// where there is a large archive this is not long enough.
	time.Sleep(time.Millisecond * 2000)

	if err := copyScripts(scripts, i.ScriptsSrcDir, appendSlash(i.OutDir)+i.ScriptsDir); err != nil {
		jww.ERROR.Println("packerTemplate.create: " + err.Error())
		return err
	}

	if err := copyDirContent(i.HTTPSrcDir, appendSlash(i.OutDir)+i.HTTPDir); err != nil {
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

	jww.TRACE.Print("Packer template directory, JSON, and contents were created and copied for " + b.BuildName)
	jww.DEBUG.Println("packerTemplate.create: exit")
	return nil
}

// copyScripts copies the passed scripts from the source to the destination.
func copyScripts(scripts []string, src string, dest string) error {
	var errCnt, okCnt int
	var wB int64
	var err error

	for _, script := range scripts {

		if wB, err = copyFile(script, src, dest); err != nil {
			jww.ERROR.Print("copyScripts: " + err.Error())
			errCnt++
		} else {
			jww.TRACE.Print("copyScripts: " + strconv.FormatInt(wB, 10) + " Bytes were copied from " + src + " to " + dest)
			okCnt++
		}

	}

	if errCnt > 0 {
		jww.ERROR.Print("copyScripts: Copy of scripts for build had " + strconv.Itoa(errCnt) + " errors. There were " + strconv.Itoa(okCnt) + " scripts that were copied without error.")
		return err
	}

	jww.TRACE.Print("copyScripts: " + strconv.Itoa(okCnt) + " scripts were successfully copied")
	return nil
}

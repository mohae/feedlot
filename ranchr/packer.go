package ranchr

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strconv"
	"time"
)

type packerTemplate struct {
	Description      string                 `json:"description"`
	MinPackerVersion string                 `json:"min_packer_version"`
	Builders         []interface{}          `json:"builders"`
	PostProcessors   []interface{}          `json:"post-processors"`
	Provisioners     []interface{}          `json:"provisioners"`
	Variables        map[string]interface{} `json:"variables"`
}

// Creates a Packer template and writes it out as a JSON file. This function
// also handles the output artifacts for the Packer template, including the 
// archival and removal of any already existing output artifacts in the same
// output directory as this Packer template.
// 
// The Dir related settings in the Rancher configuration files relate to the 
// handling of source artifacts, including their directories, and their output
// settings. 
// TODO break this up
func (p *packerTemplate) TemplateToFileJSON(i IODirInf, b BuildInf, scripts []string) error {
	logger.Debugf("%v/n%v/n%v", i, b, scripts)

	if i.HTTPDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: HTTPDir directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	if i.HTTPSrcDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: HTTPSrcDir directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	if i.OutDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: output directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	if i.SrcDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: SrcDir directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	if i.ScriptsDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: ScriptsDir directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	if i.ScriptsSrcDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: ScriptsSrcDir directory for " + b.BuildName + " not set")
		logger.Error(err.Error())
		return err
	}

	// priorBuild handles both the archiving and deletion of the prior build, if it exists, i.e.
	// if the build's output path exists.
	a := Archive{}

	if err := a.priorBuild(appendSlash(i.OutDir), "gzip"); err != nil {
		logger.Error(err.Error())
		return err
	}

	// TODO This needs to be handled better...this is too long for most builds but if there are situations
	// where there is a large archive this is not long enough.
	time.Sleep(time.Millisecond * 2000)

	var errCnt, okCnt int

	for _, script := range scripts {
		if wB, err := copyFile(i.ScriptsSrcDir, appendSlash(i.OutDir)+i.ScriptsDir, script); err != nil {
			logger.Error(err.Error())
			errCnt++
		} else {
			logger.Trace(strconv.FormatInt(wB, 10) + " Bytes were copied from " + i.ScriptsDir + script + " to " + appendSlash(i.OutDir) + script)
			okCnt++
		}
	}

	if errCnt > 0 {
		logger.Error("Copy of scripts for build, " + b.BuildName + ", had " + strconv.Itoa(errCnt) + " errors. There were " + strconv.Itoa(okCnt) + " scripts that were copied without error.")
	} else {
		logger.Trace(strconv.Itoa(okCnt) + " scripts were successfully copied for " + b.BuildName)
	}

	// Make the directory, if necessary, and copy the directory contents for the HTTP directory
	logger.Tracef("Copy HTTP directory from %s to %s", i.HTTPSrcDir, appendSlash(i.OutDir)+i.HTTPDir)

	if err := os.MkdirAll(appendSlash(i.OutDir)+i.HTTPDir, os.FileMode(0766)); err != nil {
		logger.Error(err.Error())
		return err
	}

	if err := copyDirContent(i.HTTPSrcDir, appendSlash(i.OutDir)+i.HTTPDir); err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Trace("Copied contents of " + i.HTTPSrcDir + " to " + appendSlash(i.OutDir) + i.HTTPDir)

	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	f, err := os.Create(appendSlash(i.OutDir) + b.Name + ".json")
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	// Close the file with error handling
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			logger.Error(cerr.Error())
			err = cerr
		}
	}()

	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Debug("Packer template directory, JSON, and contents were created and copied for " + b.BuildName)

	return nil
}

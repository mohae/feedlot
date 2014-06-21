package ranchr

import (
	_ "bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	_ "reflect"
	"strconv"
	_ "strings"
	_ "time"
)

type packerer interface {
	mergeSettings([]string)
}

type builderer interface {
	mergeVMSettings([]string)
}

type packerTemplate struct {
	Description      string                 `json:"description"`
	MinPackerVersion string                 `json:"min_packer_version"`
	Builders         []interface{}          `json:"builders"`
	PostProcessors   []interface{}          `json:"post-processors"`
	Provisioners     []interface{}          `json:"provisioners"`
	Variables        map[string]interface{} `json:"variables"`
}

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

	var errCnt, okCnt int
	for _, script := range scripts {

		if wB, err := copyFile(i.ScriptsSrcDir, appendSlash(i.OutDir) + i.ScriptsDir, script); err != nil {
			logger.Error(err.Error())
			errCnt++
		} else {
			logger.Info(strconv.FormatInt(wB, 10) + " Bytes were copied from " + i.ScriptsDir + script + " to " + appendSlash(i.OutDir) + script)
			okCnt++
		}
	}

	if errCnt > 0 {
		fmt.Printf("Copy of scripts for build, %s, had %s errors. There were %s scripts that were copied without error.", b.BuildName, string(errCnt), string(okCnt))
		logger.Error("Copy of scripts for build, " + b.BuildName + ", had " + strconv.Itoa(errCnt) + " errors. There were " + strconv.Itoa(okCnt) + " scripts that were copied without error.")
	} else {
		logger.Info(strconv.Itoa(okCnt) + " scripts were successfully copied.")
	}
	if err := os.MkdirAll(appendSlash(i.OutDir) +  "http", os.FileMode(0766)); err != nil {
		logger.Error(err.Error())
		return err
	}
	if err := copyDirContent(i.HTTPSrcDir, appendSlash(i.OutDir) + i.HTTPDir); err != nil {
		logger.Error(err.Error())
	}
	logger.Trace("Copied contents of " + i.HTTPSrcDir + " to " + appendSlash(i.OutDir) + i.HTTPDir)

	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		logger.Error("Marshalling of the Packer Template failed: " + err.Error())
		return err
	}

	f, err := os.Create(appendSlash(i.OutDir) + b.Name + ".json")
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	defer f.Close()

	_, err = io.WriteString(f, string(tplJSON[:]))
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("Packer template directory, JSON, and contents were created and copied for " + b.BuildName)

	return nil
}

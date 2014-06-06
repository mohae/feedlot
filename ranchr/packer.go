package ranchr


import (
	_"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	_"reflect"
	_"strings"
	_"time"
)

type packerer interface {
	mergeSettings([]string)
}

type builderer interface {
	mergeVMSettings([]string)
}

type PackerTemplate struct {
	Description      string                 `json:"description"`
	MinPackerVersion string                 `json:"min_packer_version"`
	Builders         []interface{}          `json:"builders"`
	PostProcessors   []interface{}          `json:"post-processors"`
	Provisioners     []interface{}          `json:"provisioners"`
	Variables        map[string]interface{} `json:"variables"`
}

func (p *PackerTemplate) TemplateToFileJSON(fName string, outDir string) error {
	fName = "ubuntu.JSON"
	outDir = "out/ubuntu/"

	if fName == "" {
		err := errors.New("ranchr.TemplateToFileJSON: target Packer template filename not set")
		return err
	}

	if outDir == "" {
		err := errors.New("ranchr.TemplateToFileJSON: target directory information for not set")
		return err
	}

	// Write it out as JSON
	tplJSON, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		Log.Error("Marshalling of the Packer Template failed: " + err.Error())
		return err
	}
	
	fmt.Print(string(tplJSON[:]), "\n")

	f, err := os.Create(outDir + fName)
	if err != nil {
		fmt.Println(err)
	}
	
	_, err = io.WriteString(f,  string(tplJSON[:]))
	if err != nil {
		fmt.Println(err)
	}

	f.Close()

	

	return nil
}

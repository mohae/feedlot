package ranchr

/*
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	_"reflect"
	"strings"
	"time"

	log "gopkg.in/inconshreveable/log15.v2"	
)
*/
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


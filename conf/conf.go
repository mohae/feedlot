package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohae/contour"
)

const (
	// DefaultFormat is the default configuration format.
	DefaultFormat = JSON
)

// ErrUnsupportedFormat occurs when the specified format is not supported.
var ErrUnsupportedFormat = errors.New("unsupported format")

// Feedlot setting names: these values are used in the config files, for flags,
// and as the basis for environment variables.
const (
	// ArchivePriorBuild is a boolean for whether or not a compressed tarball
	// of a prior build, for a given Packer template, should be created, if it
	// exists.
	ArchivePriorBuild = "archive_prior_build"
	// Dir is the directory that contains the Feedlot build information.
	Dir = "conf_dir"
	// Example is a bool that let's Feedlot know that the current run is an
	// example run.  Feedlot will look for the configurations and source in
	// the configured ExampleDir.
	Example = "example"
	// ExampleDir is the directory that contains Feedlot examples.  Examples
	// are used to show how Feedlot build templates may be configured and to
	// provide an easy way to generated example Packer templates.
	ExampleDir = "example_dir"
	// Format is the format used for the Feedlot configuration files: either
	// TOML or JSON.  TOML expects all configuration files to have either the
	// '.toml' or '.tml' extension.  JSON expects all configuration files to have
	// one of the following extensions: '.json', '.jsn', '.cjsn', or '.cjson'.
	// JSON is the default format.
	Format = "format"
	// ParamDelimStart is the delimiter used to indicate the start of a Feedlot
	// parameter (variable).  The default start delimiter is ':'.  This is used
	// so that Feedlot parameters in templates do not conflict with Packer
	// parameters, which use '{{ }}'.
	ParamDelimStart = "param_delim_start"
	// File is the flag name for the file to be used if logging is enabled;
	// this defaults to stderr.
	LogFile = "log_file"
	// Level is the flag name for the minimum log level used for logging.
	LogLevel = "log_level"
	// LogFlags: is the flags to use when logging: https://golang.org/pkg/log/#pkg-constants
	// In addition to the ones defined in the docs, none is also a valid value. None means
	// don't use any flags. By default, log.LstdFlags is used.
	LogFlags = "log_flags"
)

var (
	// Name is the name of the application
	Name = filepath.Base(os.Args[0])

	// File is the suffix for the ENV variable name that holds the override
	// value for the Feedlot conf file, if there is one.
	File = "conf_file"

	// Filename is the default value for the optional Feedlot conf file. This
	// may be overridden using the 'FEEDLOT_CONF_FILE' environment variable.
	Filename = "feedlot.cjson"

	// App contains the values for the loaded Feedlot configuration.
	// TODO is this still necessary?
	App app
)

// supported conf formats
const (
	UnsupportedConfFormat ConfFormat = iota
	JSON
	TOML
)

// ConfFormat: the configuration file's format.
type ConfFormat int

var confFormats = [...]string{
	"unsupported configuration format",
	"JSON",
	"TOML",
}

func (c ConfFormat) String() string { return confFormats[c] }

// ParseConfFormat returns the ConfFormat for the provided string. All values
// are lower cased prior to comparison. If a corresponding ConfFormat is not
// found, UnsupportedConfFormat is returned.
func ParseConfFormat(s string) ConfFormat {
	// make upper for consistency
	s = strings.ToUpper(s)
	switch s {
	case "JSON", "JSN", "CJSON", "CJSN":
		return JSON
	case "TOML", "TML":
		return TOML
	default:
		return UnsupportedConfFormat
	}
}

type app struct {
	ConfDir         string `toml:"conf_dir",json:"conf_dir"`
	Example         bool
	ExampleDir      string `toml:"example_dir",json:"example_dir"`
	Format          string
	LogFile         string `toml:"log_file",json:"log_file"`
	LogLevel        string `toml:"log_level",json:"log_level"`
	LogFlags        string `toml:"log_flags", json:"log_flags"`
	ParamDelimStart string `toml:"param_delim_start",json:"param_delim_start"`
}

func init() {
	// Override the defsulat cfg file if there is a value in the env var.
	name := os.Getenv(contour.GetEnvName(File))
	// if it's not set, use the application default
	if name != "" {
		Filename = name
	}

	contour.SetName(Name)
	contour.SetUseEnv(true)
	// missing main application cfg isn't considered an error state.
	contour.SetErrOnMissingCfg(false)
	contour.RegisterCfgFile(File, Filename)
	// shortcuts used: a, d, e, f, i, g, l, n, o, p, r, s, t, v, 	x
	contour.RegisterBoolFlag(ArchivePriorBuild, "v", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterStringFlag(Dir, "c", "conf/", "conf/", "location of the directory with the feedlot build configuration files")
	contour.RegisterBoolFlag(Example, "x", false, "false", "whether or not to generate from examples")
	contour.RegisterStringFlag(ExampleDir, "y", "examples/", "examples/", "location of the directory with the example feedlot build configuration files")
	contour.RegisterStringFlag(Format, "f", JSON.String(), JSON.String(), "the format of the feedlot conf files: toml or json")
	contour.RegisterStringFlag(LogFile, "g", "stderr", "stderr", "log filename")
	contour.RegisterStringFlag(LogLevel, "l", "error", "error", "log level")
	contour.RegisterStringFlag(LogFlags, "g", "", "", "'none' for no prefixes; comma separated list of log flags; default: log.LstdFlags")
	contour.RegisterStringFlag(ParamDelimStart, "p", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag("envs", "e", "", "", "additional environments from within which config additional config information should be loaded")
	contour.RegisterStringFlag("distro", "d", "", "", "specifies the distro for which a Packer template using defaults should be created")
	contour.RegisterStringFlag("arch", "a", "", "", "os arch override for default builds")
	contour.RegisterStringFlag("image", "i", "", "", "os image override for default builds")
	contour.RegisterStringFlag("release", "r", "", "", "os release override for default builds")
}

// SetAppConfFile set's the App conf from the app's conf file and then applies
// any env vars that have been set. After this, settings can only be updated
// programmatically or via command-line flags.
//
// The default conf file may not be the one found as the app conf file may be
// in a different format. SetAppConfFile first looks for it in the configured
// location. If it is not found, the alternate format is checked.
//
// Since Feedlot supports operations without a conf file, not finding one
// is not an error state.
//
// Currently supported conf file formats:
//    TOML
//    JSON || CJSN
func SetAppConfFile() error {
	// find the actual conf filename, it may have a different extension as
	// formats can have more than one accepted extension, this is mainly to
	// handle CJSON vs JSON though.  The error is ignored becuase it doesn't
	// matter if the file doesn't exist.
	fname, _, _ := ConfFilename(contour.GetString(File))
	if fname == "" {
		return nil
	}
	contour.UpdateCfgFile(File, fname)
	err := contour.SetCfg()
	if err != nil {
		err = fmt.Errorf("find configuration: %s: %s", contour.GetString(File), err)
		return err
	}
	return nil
}

// ConfFilename takea a conf file name and checks to see if it exists. If it
// doesn't exist, it checks to see if the file can be found under an alternate
// extension by checking what config format Feedlot is set to use and iterating
// through the list of supported exts for that format.  If a file exists under
// a particular file + ext combination, that is returned.  If no match is
// found, the error on the original filename is returned so that the message
// information is consistent with what is expected.
func ConfFilename(fname string) (string, ConfFormat, error) {
	cf := ParseConfFormat(contour.GetString(Format))
	_, err := os.Stat(fname)
	if err == nil {
		return fname, cf, nil
	}
	// if the file isn't found, look for it according to format extensions
	var exts []string
	switch cf {
	case JSON:
		exts = []string{"json", "jsn", "cjson", "cjsn", "JSON", "JSN", "CJSON", "CJSN"}
	case TOML:
		exts = []string{"toml", "tml", "TOML", "TML"}
	default:
		return "", UnsupportedConfFormat, fmt.Errorf("%s: unsupported conf format", contour.GetString(Format))
	}
	name := strings.TrimSuffix(fname, filepath.Ext(fname))
	for _, ext := range exts {
		n := fmt.Sprintf("%s.%s", name, ext)
		_, err := os.Stat(n)
		if err == nil {
			return n, cf, nil
		}
	}
	// nothing found
	return "", cf, fmt.Errorf("%s not found", fname)
}

// FindConfFile returns the location of the provided conf file. This accounts
// for examples. An empty
//
// If the p field has a value, it is used as the dir path, instead of the
// confDir,
func FindConfFile(p, name string) string {
	if name == "" {
		return name
	}
	var fname string
	// save the filename and add an extension to it if it doesn't exist
	if filepath.Ext(name) == "" {
		fname = name
		name = fmt.Sprintf("%s.%s", name, contour.GetString(Format))
	} else {
		fname = strings.TrimSuffix(name, filepath.Ext(name))
	}
	// if the path wasn't passed, use the confdir, unless this file is the supported
	// file. A path is prefixed to supported file only if this func receives one;
	// the ConfDir is not used for supported.
	if fname != "supported" {
		p = filepath.Join(p, contour.GetString(Dir))
	}
	if contour.GetBool(Example) {
		// example files always end in '.example'
		return filepath.Join(contour.GetString(ExampleDir), p, name)
	}
	return filepath.Join(p, name)
}

package app

import (
	"os"
	"path/filepath"

	"github.com/mohae/contour"
)

const (
	// DefaultFormat is the default configuration format.
	DefaultFormat = JSON
)

// Feedlot setting names: these values are used in the config files, for flags,
// and as the basis for environment variables.
const (
	// ArchivePriorBuild is a boolean for whether or not a compressed tarball
	// of a prior build, for a given Packer template, should be created, if it
	// exists.
	ArchivePriorBuild = "archive_prior_build"
	// ConfDir is the directory that contains the Feedlot build information.
	ConfDir = "conf_dir"
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
	// LogFile is the log file to be used if logging is enabled; this defaults to
	// stderr.
	LogFile = "log_file"
	// LogLevel is the minimum log level used for logging.
	LogLevel = "log_level"
	// Verbose prints non-log information to stdout.
	Verbose = "verbose"
)

var (
	// Name is the name of the application
	Name = filepath.Base(os.Args[0])

	// CfgFile is the suffix for the ENV variable name that holds the override
	// value for the Feedlot cfg file, if there is one.
	CfgFile = "cfg_file"

	// CfgFilename is the default value for the optional Feedlot cfg file.  This may
	// be overridden using the 'FEEDLOT_CFG_FILE' environment variable.
	CfgFilename = "feedlot.cjson"

	// AppCfg contains the values for the loaded Feedlot configuration.
	// TODO is this still necessary?
	AppCfg appCfg
)

type appCfg struct {
	ConfDir         string `toml:"conf_dir",json:"conf_dir"`
	Example         bool
	ExampleDir      string `toml:"example_dir",json:"example_dir"`
	Format          string
	LogFile         string `toml:"log_file",json:"log_file"`
	LogLevel        string `toml:"log_level",json:"log_level"`
	Verbose         string `toml:"verbose",json:"verbose"`
	ParamDelimStart string `toml:"param_delim_start",json:"param_delim_start"`
}

func init() {
	// Override the defsulat cfg file if there is a value in the env var.
	cfgFilename := os.Getenv(contour.GetEnvName(CfgFile))
	// if it's not set, use the application default
	if cfgFilename == "" {
		cfgFilename = CfgFilename

	}
	contour.SetName(Name)
	contour.SetUseEnv(true)
	// missing main application cfg isn't considered an error state.
	contour.SetErrOnMissingCfg(false)
	contour.RegisterCfgFile(CfgFile, cfgFilename)
	// shortcuts used: a, d, e, f, i, g, l, n, o, p, r, s, t, v, 	x
	contour.RegisterBoolFlag(ArchivePriorBuild, "v", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterStringFlag(ConfDir, "c", "conf/", "conf/", "location of the directory with the feedlot build configuration files")
	contour.RegisterBoolFlag(Example, "x", false, "false", "whether or not to generate from examples")
	contour.RegisterStringFlag(ExampleDir, "y", "examples/", "examples/", "location of the directory with the example feedlot build configuration files")
	contour.RegisterStringFlag(Format, "f", JSON.String(), JSON.String(), "the format of the feedlot conf files: toml or json")
	contour.RegisterStringFlag(LogFile, "g", "stderr", "stderr", "log filename")
	contour.RegisterStringFlag(LogLevel, "l", "error", "error", "log level")
	contour.RegisterBoolFlag(Verbose, "v", false, "false", "verbose output")
	contour.RegisterStringFlag(ParamDelimStart, "p", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag("envs", "e", "", "", "additional environments from within which config additional config information should be loaded")
	contour.RegisterStringFlag("distro", "d", "", "", "specifies the distro for which a Packer template using defaults should be created")
	contour.RegisterStringFlag("arch", "a", "", "", "os arch override for default builds")
	contour.RegisterStringFlag("image", "i", "", "", "os image override for default builds")
	contour.RegisterStringFlag("release", "r", "", "", "os release override for default builds")
}

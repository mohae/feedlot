package app

import (
	"os"

	"github.com/mohae/contour"
)

const (
	// Name is the name of the application
	Name = "rancher"
	// DefaultFormat is the default configuration format.
	DefaultFormat = JSON
)

// Rancher setting names: these values are used in the config files, for flags,
// and as the basis for environment variables.
const (
	// ArchivePriorBuild is a boolean for whether or not a compressed tarball
	// of a prior build, for a given Packer template, should be created, if it
	// exists.
	ArchivePriorBuild = "archive_prior_build"
	// ConfDir is the directory that contains the Rancher build information.
	ConfDir = "conf_dir"
	// Example is a bool that let's Rancher know that the current run is an
	// example run.  Rancher will look for the configurations and source in
	// the configured ExampleDir.
	Example = "example"
	// ExampleDir is the directory that contains Rancher examples.  Examples
	// are used to show how Rancher build templates may be configured and to
	// provide an easy way to generated example Packer templates.
	ExampleDir = "example_dir"
	// Format is the format used for the Rancher configuration files: either
	// TOML or JSON.  TOML expects all configuration files to have either the
	// '.toml' or '.tml' extension.  JSON expects all configuration files to have
	// one of the following extensions: '.json', '.jsn', '.cjsn', or '.cjson'.
	// JSON is the default format.
	Format = "format"
	// ParamDelimStart is the delimiter used to indicate the start of a Rancher
	// parameter (variable).  The default start delimiter is ':'.  This is used
	// so that Rancher parameters in templates do not conflict with Packer
	// parameters, which use '{{ }}'.
	ParamDelimStart = "param_delim_start"
	// Log is a bool that indicates whether Rancher should use logging.  Log
	// messages are written to both a file and stdout.
	Log = "log"
	// LogFile is the log file to be used if logging is enabled.  Rancher
	// generates a new logfile for every run.  If a file already exists with
	// the same name, Rancher will add the date and, when necessary, a sequence
	// number to ensure the file is unique.
	LogFile = "log_file"
	// LogLevelFile is the minimum log level used for logging to file.  The
	// default log level for files is 'WARN'.
	LogLevelFile = "log_level_file"
	// LogLevelStdOut is the minimum log level used for printing log messages
	// to stdout.  The default log level for stdout is 'ERROR'.
	LogLevelStdOut = "log_level_stdout"
)

// CfgFile is the suffix for the ENV variable name that holds the override
// value for the Rancher cfg file, if there is one.
var CfgFile = "cfg_file"

// CfgFilename is the default value for the optional Rancher cfg file.  This may
// be overridden using the 'RANCHER_CFG_FILE' environment variable.
var CfgFilename = "rancher.json"

// AppCfg contains the values for the loaded Rancher configuration.
// TODO is this still necessary?
var AppCfg appCfg

type appCfg struct {
	ConfDir         string `toml:"conf_dir",json:"conf_dir"`
	Example         bool
	ExampleDir      string `toml:"example_dir",json:"example_dir"`
	Format          string
	Log             bool
	LogFile         string `toml:"log_file",json:"log_file"`
	LogLevelFile    string `toml:"log_level_file",json:"log_level_file"`
	LogLevelStdout  string `toml:"log_level_stdout",json:"log_level_stdout"`
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
	contour.RegisterBoolFlag(Log, "l", false, "false", "enable/disable logging")
	contour.RegisterStringFlag(ConfDir, "c", "conf/", "conf/", "location of the directory with the rancher build configuration files")
	contour.RegisterBoolFlag(Example, "e", false, "false", "whether or not to generate from examples")
	contour.RegisterStringFlag(ExampleDir, "x", "examples/", "examples/", "location of the directory with the example rancher build configuration files")
	contour.RegisterStringFlag(Format, "t", JSON.String(), JSON.String(), "the format of the Rancher conf files: toml or json")
	contour.RegisterStringFlag(LogFile, "g", "rancher.log", "rancher.log", "log filename")
	contour.RegisterStringFlag(LogLevelFile, "f", "WARN", "WARN", "log level for writing to the log file")
	contour.RegisterStringFlag(LogLevelStdOut, "o", "ERROR", "ERROR", "log level for writing to stdout")
	contour.RegisterStringFlag(ParamDelimStart, "p", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag("envs", "n", "", "", "additional environments from within which config additional config information should be loaded")
	contour.RegisterStringFlag("distro", "d", "", "", "specifies the distro for which a Packer template using defaults should be created")
	contour.RegisterStringFlag("arch", "a", "", "", "os arch override for default builds")
	contour.RegisterStringFlag("image", "i", "", "", "os image override for default builds")
	contour.RegisterStringFlag("release", "r", "", "", "os release override for default builds")
}

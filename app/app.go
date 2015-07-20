package app

import (
	"os"

	"github.com/mohae/contour"
)

const (
	Name          = "rancher"
	DefaultFormat = "json"
)

// Rancher conf filename roots; these are reserved names for conf files.
const (
	Build     = "build"
	BuildList = "build_list"
	Default   = "default"
	Supported = "supported"
)

// Rancher cfg and flag setting names
const (
	ArchivePriorBuild = "archive_prior_build"
	ConfDir           = "conf_dir"
	ExampleDir        = "example_dir"
	SourceDir         = "source_dir"
	Format            = "format"
	Example           = "example"
	ParamDelimStart   = "param_delim_start"
	Log               = "log"
	LogFile           = "log_file"
	LogLevelFile      = "log_level_file"
	LogLevelStdOut    = "log_level_stdout"
)

// ExampleExt is the extension for example files; used in example mode.
const ExampleExt = ".example"

// CfgFile is the suffix for the ENV variable name that holds the override
// value for the Rancher cfg file, if there is one.
var CfgFile = "cfg_file"

// CfgFilename is the default value for the optional Rancher cfg file.
var CfgFilename = "rancher.toml"

// AppCfg contains the current Rancher cfguration...loaded at start-up.
var AppCfg appCfg

type appCfg struct {
	ConfDir         string `toml:"conf_dir",json:"conf_dir"`
	ExampleDir      string `toml:"example_dir",json:"example_dir"`
	Format          string
	ParamDelimStart string `toml:"param_delim_start",json:"param_delim_start"`
	Example         bool
	Log             bool
	LogFile         string `toml:"log_file",json:"log_file"`
	LogLevelFile    string `toml:"log_level_file",json:log_level_file"`
	LogLevelStdout  string `toml:"log_level_stdout",json:"log_level_stdout"`
}

// ArgsFilter has all the valid commandline flags for the build-subcommand.
type ArgsFilter struct {
	// Arch is a distribution specific string for the OS's target
	// architecture.
	Arch string
	// Distro is the name of the distribution, this value is consistent
	// with Packer.
	Distro string
	// Image is the type of ISO image that is to be used. This is a
	// distribution specific value.
	Image string
	// Release is the release number or string of the ISO that is to be
	// used. The valid values are distribution specific.
	Release string
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
	contour.RegisterStringFlag(Format, "", "toml", "toml", "the format of the Rancher conf files: toml or json")
	// shortcuts used: a, d, e, eg, f, i, l, n, p, r, s, v
	contour.RegisterBoolFlag(ArchivePriorBuild, "v", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterBoolFlag(Log, "l", false, "false", "enable/disable logging")
	contour.RegisterBoolFlag(Example, "eg", false, "false", "whether or not to generate from examples")
	contour.RegisterStringFlag(ConfDir, "", "conf/", "conf/", "location of the root configuration directory for conf")
	contour.RegisterStringFlag(ExampleDir, "", "examples/", "examples/", "the location of the root directory for example rancher configuration files")
	contour.RegisterStringFlag(ParamDelimStart, "p", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag(LogFile, "", "rancher.log", "rancher.log", "log filename")
	contour.RegisterStringFlag(LogLevelFile, "f", "WARN", "WARN", "log level for writing to the log file")
	contour.RegisterStringFlag(LogLevelStdOut, "s", "ERROR", "ERROR", "log level for writing to stdout")
	contour.RegisterStringFlag("envs", "e", "", "", "additional environments from within which config additional config information should be loaded")
	contour.RegisterStringFlag("distro", "d", "", "", "distro override for default builds")
	contour.RegisterStringFlag("arch", "a", "", "", "os arch override for default builds")
	contour.RegisterStringFlag("image", "i", "", "", "os image override for default builds")
	contour.RegisterStringFlag("release", "r", "", "", "os release override for default builds")
}

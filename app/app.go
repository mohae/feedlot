package app

import (
	"os"
	"strings"

	"github.com/mohae/contour"
)

const (
	Name            = "rancher"
	Build           = "build"
	BuildList       = "build_list"
	Default         = "default"
	DefaultFormat   = "json"
	Supported       = "supported"
	CfgFile         = "cfg_file"
	CfgFilename     = "rancher.json"
	ParamDelimStart = "param_delim_start"

	Log            = "log"
	LogFile        = "log_file"
	LogLevelFile   = "log_level_file"
	LogLevelStdOut = "log_level_stdout"
)

// AppCfg contains the current Rancher cfguration...loaded at start-up.
var AppCfg appCfg

type appCfg struct {
	ConfDir         string `toml:"conf_dir",json:"conf_dir"`
	ExampleDir      string `toml:"example_dir",json:"example_dir"`
	SrcDir          string `toml:"src_dir",json:"src_dir"`
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
	cfgFilename := os.Getenv(contour.GetEnvName(CfgFile))
	// if it's not set, use the application default
	if cfgFilename == "" {
		cfgFilename = CfgFilename

	}
	parts := strings.Split(cfgFilename, ".")
	if len(parts) < 2 {
		contour.RegisterStringFlag("format", "f", "toml", "toml", "the format of the rancher conf files: toml and json are supported")
	} else {
		contour.RegisterStringFlag("format", "f", parts[len(parts)-1], parts[len(parts)-1], "the format of the rancher conf files: toml and json are supported")
	}
	contour.SetName(Name)
	contour.SetUseEnv(true)
	// missing main application cfg isn't considered an error state.
	contour.SetErrOnMissingCfg(false)
	contour.RegisterCfgFile(CfgFile, CfgFilename)
	// shortcuts used: a, d, e, eg, f, i, l, n, p, r, s, v
	contour.RegisterBoolFlag("archive_prior_build", "v", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterBoolFlag(Log, "l", true, "true", "enable/disable logging")
	contour.RegisterBoolFlag("example", "eg", false, "false", "whether or not to generate from examples")
	contour.RegisterStringFlag("conf_dir", "", "conf/", "conf/", "location of the root configuration directory for conf")
	contour.RegisterStringFlag("example_dir", "", "examples/", "examples/", "the location of the root directory for example rancher configuration files")
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

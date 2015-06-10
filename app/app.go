package app

import (
	"os"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

const (
	Name            = "rancher"
	BuildFile       = "build_file"
	BuildListFile   = "build_list_file"
	CfgFile         = "cfg_file"
	CfgFilename     = "rancher.toml"
	DefaultFile     = "default_file"
	Log             = "log"
	LogFile         = "log_file"
	LogLevelFile    = "log_level_file"
	LogLevelStdOut  = "log_level_stdout"
	ParamDelimStart = "param_delim_start"
	SupportedFile   = "supported_file"
)

// AppCfg contains the current Rancher cfguration...loaded at start-up.
var AppCfg appCfg

type appCfg struct {
	BuildFile       string `toml:"build_file"`
	BuildListFile   string `toml:"build_list_file"`
	DefaultFile     string `toml:"default_file"`
	Log             bool   `toml:"log"`
	LogFile         string `toml:"log_file"`
	LogLevelFile    string `toml:"log_level_file"`
	LogLevelStdout  string `toml:"log_level_stdout"`
	ParamDelimStart string `toml:"param_delim_start"`
	SupportedFile   string `toml:"supported_file"`
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
	contour.SetName(Name)
	contour.SetUseEnv(true)
	// missing main application cfg isn't considered an error state.
	contour.SetErrOnMissingCfg(false)
	contour.RegisterCfgFile(CfgFile, CfgFilename)
	// shortcuts used: a, d, f, i, l, n, p, r, s, v
	contour.RegisterBoolFlag("archive_prior_build", "v", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterBoolFlag(Log, "l", true, "true", "enable/disable logging")
	contour.RegisterStringFlag(BuildFile, "", "conf.d/build.toml", "conf.d/build.toml", "location of the build cfguration file")
	contour.RegisterStringFlag(BuildListFile, "", "conf.d/build_list.toml", "conf.d/build_list.toml", "location of the build list cfguration file")
	contour.RegisterStringFlag(DefaultFile, "", "conf/default.toml", "conf/default.toml", "location of the default cfguration file")
	contour.RegisterStringFlag(SupportedFile, "", "conf/supported.toml", "conf/supported.toml", "location of the supported distros cfguration file")
	contour.RegisterStringFlag(ParamDelimStart, "p", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag(LogFile, "", "rancher.log", "rancher.log", "log filename")
	contour.RegisterStringFlag(LogLevelFile, "f", "WARN", "WARN", "log level for writing to the log file")
	contour.RegisterStringFlag(LogLevelStdOut, "s", "ERROR", "ERROR", "log level for writing to stdout")
	contour.RegisterStringFlag("distro", "d", "", "", "distro override for default builds")
	contour.RegisterStringFlag("arch", "a", "", "", "os arch override for default builds")
	contour.RegisterStringFlag("image", "i", "", "", "os image override for default builds")
	contour.RegisterStringFlag("release", "r", "", "", "os release override for default builds")
}

// SetCfg set's the appCFg from the app's cfg file and then applies any
// env vars that have been set. After this, settings can only be updated
// programmatically or via command-line flags.
func SetCfg() error {
	err := contour.SetCfg()
	if err != nil {
		jww.ERROR.Print(err)
		jww.FEEDBACK.Printf("SetCfg: %s", err.Error())
		return err
	}
	return nil
}

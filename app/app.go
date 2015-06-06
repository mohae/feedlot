package app

import (
	"os"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

const (
	Name               = "rancher"
	CfgFilename string = "rancher.toml"
)

// Constants for names of configuration options. Constants that end in Flag are
// not available as Envorinment variables. All of the following constants are
// exposed as flags too.
const (
	ArchFlag          = "arch"
	ArchivePriorBuild = "rancher_archivepriorbuild"
	BuildFile         = "rancher_buildfile"
	BuildListFile     = "rancher_buildlistfile"
	CfgFile           = "rancher_cfgfile"
	DefaultFile       = "rancher_defaultFile"
	DistroFlag        = "distro"
	ImageFlag         = "image"
	Log               = "rancher_log"
	LogFile           = "rancher_logfile"
	LogLevelFile      = "rancher_loglevelfile"
	LogLevelStdOut    = "rancher_loglevelstdout"
	ParamDelimStart   = "rancher_paramdelimstart"
	ReleaseFlag       = "release"
	SupportedFile     = "rancher_supportedfile"
)

// AppConfig contains the current Rancher configuration...loaded at start-up.
var AppConfig appConfig

type appConfig struct {
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
	cfgFilename := os.Getenv(CfgFile)
	// if it's not set, use the application default
	if cfgFilename == "" {
		cfgFilename = CfgFilename
	}
	contour.RegisterCfgFile(CfgFile, cfgFilename)
	contour.RegisterStringCore("name", Name)
	contour.RegisterBoolFlag(ArchivePriorBuild, "v", "rancher_prior_build", false, "false", "archive prior build before writing new packer template files")
	contour.RegisterBoolFlag(Log, "l", true, "true", "rancher_log", "enable/disable logging")
	contour.RegisterStringFlag(BuildFile, "", "rancher_build_file". "conf.d/build.toml", "conf.d/build.toml", "location of the build configuration file")
	contour.RegisterStringFlag(BuildListFile, "", "rancher_build_list_file", "conf.d/build_list.toml", "conf.d/build_list.toml", "location of the build list configuration file")
	contour.RegisterStringFlag(DefaultFile, "", "rancher_default_file", "conf/default.toml", "conf/default.toml", "location of the default configuration file")
	contour.RegisterStringFlag(SupportedFile, "", "rancher_supported_file", "conf/supported.toml", "conf/supported.toml", "location of the supported distros configuration file")
	contour.RegisterStringFlag(ParamDelimStart, "p", "rancher_param_delim_start", ":", ":", "the start delimiter for template variabes")
	contour.RegisterStringFlag(LogFile, "n", "rancher_log_file", "rancher.log", "rancher.log", "log filename")
	contour.RegisterStringFlag(LogLevelFile, "f", "rancher_log_level_file", "WARN", "WARN", "log level for writing to the log file")
	contour.RegisterStringFlag(LogLevelStdOut, "s", "rancher_log_level_std_out", "ERROR", "ERROR", "log level for writing to stdout")
	contour.RegisterStringFlag(DistroFlag, "d", "", "", "", "distro override for default builds")
	contour.RegisterStringFlag(ArchFlag, "a", "", "", "", "os arch override for default builds")
	contour.RegisterStringFlag(ImageFlag, "i", "", "", "", "os image override for default builds")
	contour.RegisterStringFlag(ReleaseFlag, "r", "", "", "", "os release override for default builds")
}

// After this, only overrides can occur via command flags.
func SetCfg() error {
	err := contour.SetCfg()
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	return nil
}

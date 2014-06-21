package ranchr

import (
	"errors"
	"fmt"
	"os"
	_ "reflect"
	"strings"
	"time"
)

type rawTemplate struct {
	PackerInf
	IODirInf
	BuildInf
	date    string // ISO 8601 Date format
	delim   string
	Type    string
	Arch    string
	Image   string
	Release string
	varVals map[string]string
	vars    map[string]string
	//	releaseInfo isoer{}
	build
}

// Returns a rawTemplate with current date in ISO 8601 format. This should be
// called when a rawTemplate with the current date is desired.
func newRawTemplate() rawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	r := rawTemplate{date: splitDate[0], delim: os.Getenv(EnvParamDelimStart)}

	return r
}

//
func (r *rawTemplate) createDistroTemplate(d rawTemplate) {
	r.IODirInf = d.IODirInf
	r.PackerInf = d.PackerInf
	r.BuildInf = d.BuildInf
	r.Arch = d.Arch
	r.BaseURL = d.BaseURL
	r.Type = d.Type
	r.Image = d.Image
	r.Release = d.Release
	r.BuilderType = d.BuilderType
	r.Builders = d.Builders
	r.Provisioners = d.Provisioners
	r.PostProcessors = d.PostProcessors
	logger.Tracef("DistroTemplate: %+v", r)
	return
}

// Create a Packer template from the rawTemplate that can be marshalled to JSON.
func (r *rawTemplate) createPackerTemplate() (packerTemplate, error) {
	logger.Debugf("Current rawTemplate state is: %+v", r)
	var err error
	var vars map[string]interface{}
	r.mergeVariables()
	// General Packer Stuff
	p := packerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description

	// Builders
	iSl := make([]interface{}, len(r.Builders))

	if p.Builders, vars, err = r.createBuilders(); err != nil {
		logger.Error(err.Error())
		return p, err
	}
	// Post-Processors
	var i int
	var sM map[string]interface{}
	iSl = make([]interface{}, len(r.PostProcessors))

	for k, pp := range r.PostProcessors {
		sM = pp.settingsToMap(k, r)

		if sM == nil {
			err = errors.New("an error occured while trying to create the Packer post-processor template for " + k)
			logger.Critical(err.Error())
			return p, err
		}

		iSl[i] = sM
		i++
	}

	p.PostProcessors = iSl

	// Provisioners
	i = 0
	iM := make(map[string]interface{})
	iSl = make([]interface{}, len(r.Provisioners))
	for k, pr := range r.Provisioners {
		iM = pr.settingsToMap(k, r)
		if iM == nil {
			err = errors.New("CreatePackerTemplate error: the Settings map is nil")
			logger.Critical(err.Error())
			return p, err
		}

		// If there are any scripts add them. Scripts are already in an array. Scripts use
		// a map[string]interface{} structure for JSON
		if len(pr.Scripts) > 0 {

			// The variables in each script element need to be replaced.
			for i, script := range pr.Scripts {
				pr.Scripts[i] = r.replaceVariables(script)
			}

			iM["scripts"] = pr.Scripts
		}
		iSl[i] = iM
		i++
	}

	p.Provisioners = iSl
	p.Variables = vars

	// Now we can create the Variable Section

	// Return the generated Packer Template
	logger.Info("PackerTemplate created from a rawTemplate.")

	return p, nil
}

func (r *rawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	// Takes a raw builder and create the appropriate Packer Builders along with a
	// slice of variables for that section builder type. Some Settings are in-lined
	// instead of adding them to the variable section.

	if r.BuilderType == nil || len(r.BuilderType) <= 0 {
		err = fmt.Errorf("no builder types were configured, unable to create builders")
		logger.Error(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var k, val, v string
	var i, ndx int
	bldrs = make([]interface{}, len(r.BuilderType))
	for _, bType := range r.BuilderType {
		logger.Debug(bType)
		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWare:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				logger.Error(err.Error())
				return nil, nil, err
			}
			tmpS["type"] = bType

			// Generate builder specific section
			tmpvm := make(map[string]string, len(r.Builders[bType].VMSettings))

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				switch k {
				case "memory":
					//do nothing
				default:
					val = r.replaceVariables(val)
					tmpvm[k] = val
					tmpS["vmx_data"] = tmpvm
				}
			}

		case BuilderVBox:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				logger.Error(err.Error())
				return nil, nil, err
			}

			tmpS["type"] = bType
			// Generate Packer Variables
			// Generate builder specific section
			tmpVB := make([][]string, len(r.Builders[bType].VMSettings))
			ndx = 0

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				switch k {
				case "memory":
					// do nothing
				default:
					val = r.replaceVariables(val)
					tmpVB[i] = make([]string, 4)
					tmpVB[i][0] = "modifyvm"
					tmpVB[i][1] = "{{.Name}}"
					tmpVB[i][2] = "--" + k
					tmpVB[i][3] = val
				}
			}
			tmpS["vboxmanage"] = tmpVB

		default:
			err = errors.New("the requested builder, '" + bType + "', is not supported")
			logger.Error(err.Error())
			return nil, nil, err
		}
		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return bldrs, vars, nil
}

func (r *rawTemplate) replaceVariables(s string) string {
	// Checks incoMing string for variables and replaces them with their values.

	//see if the delim is in the string
	if strings.Index(s, r.delim) < 0 {
		return s
	}

	for vName, vVal := range r.varVals {
		s = strings.Replace(s, vName, vVal, -1)
	}

	return s
}

func (r *rawTemplate) variableSection() (map[string]interface{}, error) {
	// Generates the variable section.
	var v map[string]interface{}
	v = make(map[string]interface{})
	return v, nil
}

func (r *rawTemplate) commonVMSettings(builderType string, old []string, new []string) (Settings map[string]interface{}, vars []string, err error) {
	// Generates the common builder sections for vmWare and VBox
	var k, v string
	var tmpSl []string
	maxLen := len(old) + len(new) + 4
	tmpSl = make([]string, maxLen)
	//	vars = make([]string, maxLen)
	Settings = make(map[string]interface{}, maxLen)
	tmpSl = mergeSettingsSlices(old, new)

	for _, s := range tmpSl {
		//		var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)

		switch k {
		case "iso_checksum_type":
			// look up the release information and
			// add all the iso entries to the map

			var notSupported string
			Settings[k] = v
			switch r.Type {
			case "ubuntu":
				rel := &ubuntu{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: strings.ToUpper(v)}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
				rel.SetISOInfo()
				Settings["iso_url"] = rel.URL
				Settings["iso_checksum"] = rel.Checksum
			case "centos":
				rel := &centOS{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: strings.ToLower(v)}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
				rel.SetISOInfo()
				Settings["iso_url"] = rel.URL
				Settings["iso_checksum"] = rel.Checksum
			default:
				notSupported = r.Type + " is not supported"
			}

			if notSupported != "" {
				Settings["iso_url"] = notSupported
				Settings["iso_checksum"] = notSupported
			}
		case "boot_command", "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string
			if commands, err = commandsFromFile(v); err != nil {
				logger.Warn("error", err.Error())
				Settings[k] = "Error: " + err.Error()
				err = nil
			} else {
				// Boot commands are slices, the others are just a string.
				if k == "boot_command" {
					Settings[k] = commands
				} else {
					// Assume it's the first element.
					Settings[k] = commands[0]
				}
			}
		case "guest_os_type":
			switch r.Type {
			case "ubuntu":
				rel := &ubuntu{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: strings.ToLower(v)}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
				Settings[k] = rel.getOSType(builderType)
			case "centos":
				rel := &centOS{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: strings.ToLower(v)}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
				rel.setBaseURL()
				Settings[k] = rel.getOSType(builderType)
			}
		default:
			// just use the value
			Settings[k] = v
		}
		//		kname = getVariableName(k)

	}

	return Settings, vars, nil
}

func (r *rawTemplate) mergeBuildSettings(bld rawTemplate) {
	logger.Debug(bld)
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(bld.IODirInf)
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)

	// If defined, Builders override any prior builder Settings.
	if bld.BuilderType != nil && len(bld.BuilderType) > 0 {
		r.BuilderType = bld.BuilderType
	}

	// merge the build portions.
	r.Builders = getMergedBuilders(r.Builders, bld.Builders)
	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, bld.PostProcessors)
	r.Provisioners = getMergedProvisioners(r.Provisioners, bld.Provisioners)

	return
}

func (r *rawTemplate) mergeDistroSettings(d distro) {
	logger.Debugf("%v\n%v", r, d)
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)

	r.BuildInf.update(d.BuildInf)
	// If defined, Builders override any prior builder Settings
	if d.BuilderType != nil && len(d.BuilderType) > 0 {
		r.BuilderType = d.BuilderType
	}

	// merge the build portions.
	logger.Debugf("Merging old Builder: %v\nand\nnew Builder: %v", r.Builders, d.Builders)
	r.Builders = getMergedBuilders(r.Builders, d.Builders)
	logger.Debugf("Merged Builder: %v", r.Builders)
	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, d.PostProcessors)
	r.Provisioners = getMergedProvisioners(r.Provisioners, d.Provisioners)

	return
}

// Get a slice of script names from the shell provisioner, if any.
func (r *rawTemplate) ScriptNames() []string {
	logger.Debug("entering")
	var s []string

	if len(r.Provisioners["shell"].Scripts) > 0 {
		s = make([]string, len(r.Provisioners["shell"].Scripts))

		for i, script := range r.Provisioners["shell"].Scripts {
			//explode on "/"
			parts := strings.Split(script, "/")

			// the last element is the script name
			s[i] = parts[len(parts)-1]
		}

	}
	logger.Debug(s)
	return s

}

func (r *rawTemplate) mergeVariables() {
	// Set the src_dir and out_dir, in case there are variables embedded in them.
	// These can be embedded in other dynamic variables so they need to be resolved
	// first to avoid a mutation issue. Only Rancher static variables can be used
	// in these two Settings.
	// check src_dir and out_dir first:
	// TODO: replase this mess with something cleaner/resilient

	// Get the delim and set the replacement map, resolve name information
	r.varVals = map[string]string{r.delim + "type": r.Type, r.delim + "release": r.Release, r.delim + "arch": r.Arch, r.delim + "image": r.Image, r.delim + "date": r.date, r.delim + "build_name": r.BuildName}

	r.Name = r.replaceVariables(r.Name)

	// Src and Outdir are next, since they can be embedded too
	r.varVals[r.delim+"name"] = r.Name

	r.SrcDir = trimSuffix(r.replaceVariables(r.SrcDir), "/")
	r.OutDir = trimSuffix(r.replaceVariables(r.OutDir), "/")
	// Commands and scripts dir need to be resolved next
	r.varVals[r.delim+"out_dir"] = r.OutDir
	r.varVals[r.delim+"src_dir"] = r.SrcDir

	r.CommandsSrcDir = trimSuffix(r.replaceVariables(r.CommandsSrcDir), "/")
	r.HTTPDir = trimSuffix(r.replaceVariables(r.HTTPDir), "/")
	r.HTTPSrcDir = trimSuffix(r.replaceVariables(r.HTTPSrcDir), "/")
	r.OutDir = trimSuffix(r.replaceVariables(r.OutDir), "/")
	r.ScriptsDir = trimSuffix(r.replaceVariables(r.ScriptsDir), "/")
	r.ScriptsSrcDir = trimSuffix(r.replaceVariables(r.ScriptsSrcDir), "/")
	r.SrcDir = trimSuffix(r.replaceVariables(r.SrcDir), "/")

	// Create a full variable replacement map, know that the SrcDir and OutDir stuff are resolved.
	// Rest of the replacements are done by the packerers.
	r.varVals[r.delim+"commands_src_dir"] = r.CommandsSrcDir
	r.varVals[r.delim+"http_dir"] = r.HTTPDir
	r.varVals[r.delim+"http_src_dir"] = r.HTTPSrcDir
	r.varVals[r.delim+"out_dir"] = r.OutDir
	r.varVals[r.delim+"scripts_dir"] = r.ScriptsDir
	r.varVals[r.delim+"scripts_src_dir"] = r.ScriptsSrcDir
	r.varVals[r.delim+"src_dir"] = r.SrcDir

	r.CommandsSrcDir = trimSuffix(r.replaceVariables(r.CommandsSrcDir), "/")
	r.HTTPDir = trimSuffix(r.replaceVariables(r.HTTPDir), "/")
	r.HTTPSrcDir = trimSuffix(r.replaceVariables(r.HTTPSrcDir), "/")
	r.OutDir = trimSuffix(r.replaceVariables(r.OutDir), "/")
	r.ScriptsDir = trimSuffix(r.replaceVariables(r.ScriptsDir), "/")
	r.ScriptsSrcDir = trimSuffix(r.replaceVariables(r.ScriptsSrcDir), "/")
	r.SrcDir = trimSuffix(r.replaceVariables(r.SrcDir), "/")
}


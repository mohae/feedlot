package ranchr

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

type rawTemplate struct {
	PackerInf
	IODirInf
	BuildInf

	// holds release information
	releaseISO releaser

	// the builder specific string for the template's OS and Arch
	osType string

	// Convenience for distros that need ChecksumType for more than finding
	// the ISO checksum, e.g. CentOS.
	ChecksumType string

	// Current date in ISO 8601
	date string

	// The character(s) used to identify variables for Rancher. By default
	// this is a colon, :. Currently only a starting delimeter is supported.
	delim string

	// The distro that this template targets. The type must be a supported
	// type, i.e. defined in supported.toml. The values for type are
	// consistent with Packer values.
	Type string

	// The architecture for the ISO image, this is either 32bit or 64bit,
	// with the actual values being dependent on the operating system and
	// the target builder.
	Arch string

	// The image for the ISO. This is distro dependent.
	Image string

	// The release, or version, for the ISO. Usage and values are distro
	// dependent, however only version currently supported images that are
	// available on the distro's download site are supported.
	Release string

	// Values for variables...currently not supported
	varVals map[string]string

	// Variable name mapping...currently not supported
	vars map[string]string

	// Contains all the build information needed to create the target Packer
	// template and its associated artifacts.
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

// Assign the default template values to the rawTemplate.
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
	return
}

// Create a Packer template from the rawTemplate that can be marshalled to JSON.
func (r *rawTemplate) createPackerTemplate() (packerTemplate, error) {
	jww.DEBUG.Printf("Entering...")
	var err error
	var vars map[string]interface{}

	// Resolve the Rancher variables to their final values.
	r.mergeVariables()

	// General Packer Stuff
	p := packerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description

	// Builders
	iSl := make([]interface{}, len(r.Builders))
	if p.Builders, vars, err = r.createBuilders(); err != nil {
		jww.ERROR.Println(err.Error())
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
			jww.ERROR.Println(err.Error())
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
		iM, err = pr.settingsToMap(k, r)
		if err != nil {
			jww.ERROR.Println(err.Error())
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
	// TODO

	// Return the generated Packer Template
	jww.DEBUG.Println("PackerTemplate created from a rawTemplate.")

	return p, nil
}

// Takes a raw builder and create the appropriate Packer Builders along with a
// slice of variables for that section builder type. Some Settings are in-lined
// instead of adding them to the variable section.
func (r *rawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	if r.BuilderType == nil || len(r.BuilderType) <= 0 {
		err = fmt.Errorf("no builder types were configured, unable to create builders")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var k, val, v string
	var i, ndx int
	bldrs = make([]interface{}, len(r.BuilderType))

	// Generate the builders for each builder type.
	for _, bType := range r.BuilderType {
		jww.TRACE.Println(bType)
		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWare:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			tmpS["type"] = bType

			// Generate builder specific section
			tmpvm := make(map[string]string, len(r.Builders[bType].VMSettings))

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				val = r.replaceVariables(val)
				tmpvm[k] = val
				tmpS["vmx_data"] = tmpvm
			}

		case BuilderVBox:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			tmpS["type"] = bType
			// Generate Packer Variables
			// Generate builder specific section
			tmpVB := make([][]string, len(r.Builders[bType].VMSettings))
			ndx = 0

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				val = r.replaceVariables(val)
				tmpVB[i] = make([]string, 4)
				tmpVB[i][0] = "modifyvm"
				tmpVB[i][1] = "{{.Name}}"
				tmpVB[i][2] = "--" + k
				tmpVB[i][3] = val
			}
			tmpS["vboxmanage"] = tmpVB

		default:
			err = errors.New("the requested builder, '" + bType + "', is not supported")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}
		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return bldrs, vars, nil
}

// Checks incoming string for variables and replaces them with their values.
func (r *rawTemplate) replaceVariables(s string) string {
	//see if the delim is in the string
	if strings.Index(s, r.delim) < 0 {
		return s
	}

	// Go through each variable and replace as applicable.
	for vName, vVal := range r.varVals {
		s = strings.Replace(s, vName, vVal, -1)
	}

	return s
}

// Generates the variable section.
func (r *rawTemplate) variableSection() (map[string]interface{}, error) {
	var v map[string]interface{}
	v = make(map[string]interface{})
	return v, nil
}

// Generates the common builder sections for vmWare and VBox
func (r *rawTemplate) commonVMSettings(builderType string, old []string, new []string) (Settings map[string]interface{}, vars []string, err error) {
	var k, v string
	var mergedSlice []string

	maxLen := len(old) + len(new) + 4
	mergedSlice = make([]string, maxLen)
	Settings = make(map[string]interface{}, maxLen)
	mergedSlice = mergeSettingsSlices(old, new)

	// First set the ISO info for the desired release, if it's not already set
	if r.osType == "" {
		err = r.ISOInfo(builderType, mergedSlice)
		if err != nil {
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}
	}
	jww.TRACE.Printf("rawTemplate post r.ISOInfo(): %v", r)
	for _, s := range mergedSlice {
		//		var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		jww.TRACE.Printf("%s:%s", k, v)
		switch k {
		case "iso_checksum_type":
			switch r.Type {
			case "ubuntu":
				Settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
				Settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum

			case "centos":
				Settings["iso_url"] = r.releaseISO.(*centOS).isoURL
				Settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum

			case "default":
				err = errors.New("rawTemplate.CommonVMSettings: " + k + " is not a supported builder type")
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

		case "boot_command", "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			} 
			
			// Boot commands are slices, the others are just a string.
			if k == "boot_command" {
				Settings[k] = commands
			} else {
				// Assume it's the first element.
				Settings[k] = commands[0]
			}

		case "guest_os_type":
			Settings[k] = r.osType

		default:
			// just use the value
			Settings[k] = v
		}
	}
	return Settings, vars, nil
}

// merges Settings between an old and new template.
// Note: Arch, Image, and Release are not updated here as how these fields
// are updated depends on whether this is a build from a distribution's
// default template or from a defined build template.
func (r *rawTemplate) mergeBuildSettings(bld rawTemplate) {
	jww.DEBUG.Print(bld)
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

func (r *rawTemplate) mergeDistroSettings(d *distro) {
	jww.TRACE.Printf("%v\n%v", r, d)
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
	jww.TRACE.Printf("Merging old Builder: %v\nand new Builder: %v", r.Builders, d.Builders)
	r.Builders = getMergedBuilders(r.Builders, d.Builders)
	jww.TRACE.Printf("Merged Builder: %v", r.Builders)
	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, d.PostProcessors)
	r.Provisioners = getMergedProvisioners(r.Provisioners, d.Provisioners)

	return
}

// Get a slice of script names from the shell provisioner, if any.
func (r *rawTemplate) ScriptNames() []string {
	jww.DEBUG.Println("Entering rawTemplate.ScriptNames...")
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
	jww.DEBUG.Println("Exiting rawTemplate.ScriptNames...")
	return s

}

// Set the src_dir and out_dir, in case there are variables embedded in them.
// These can be embedded in other dynamic variables so they need to be resolved
// first to avoid a mutation issue. Only Rancher static variables can be used
// in these two Settings.
func (r *rawTemplate) mergeVariables() {
	// check src_dir and out_dir first:
	// TODO: replace this mess with something cleaner/resilient because this
	// is ugly and poorly written code...My Bad :(

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

func (r *rawTemplate) ISOInfo(builderType string, settings []string) error {
	jww.TRACE.Printf("Entering rawTemplate.ISOInfo")
	var k, v, checksumType string
	var err error

	// Only the iso_checksum_type is needed for this.
	for _, s := range settings {
		k, v = parseVar(s)

		switch k {
		case "iso_checksum_type":
			checksumType = v
		}
	}

	switch r.Type {
	case "ubuntu":
		r.releaseISO = &ubuntu{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: checksumType}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
		r.releaseISO.SetISOInfo()

		r.osType, err = r.releaseISO.(*ubuntu).getOSType(builderType)
		if err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	case "centos":
		r.releaseISO = &centOS{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: checksumType}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}
		r.releaseISO.SetISOInfo()

		r.osType, err = r.releaseISO.(*centOS).getOSType(builderType)
		if err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}
	jww.TRACE.Printf("Exit rawTemplate.ISOInfo")
	return nil
}

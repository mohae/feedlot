package ranchr

import (
	"errors"
	"fmt"
	"os"
	_ "reflect"
	"strings"
	"time"

//	log "gopkg.in/inconshreveable/log15.v2"
)

type RawTemplate struct {
	PackerInf
	IODirInf
	date		string	// ISO 8601 Date format
	BaseURL	string
	delim 	string
	Type    string
	name    string
	Arch    string
	Image   string
	Release string
	varVals		map[string]string
	vars		map[string]string
	build
}

// Returns a RawTemplate with current date in ISO 8601 format. This should be
// called when a RawTemplate with the current date is desired.
func newRawTemplate() RawTemplate	{
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	R := RawTemplate{date: splitDate[0], delim: os.Getenv(EnvParamDelimStart)}

	return R
}

// 
func (r *RawTemplate) createDistroTemplate(d RawTemplate) {
	r.IODirInf = d.IODirInf
	r.PackerInf = d.PackerInf
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

// Create a Packer template from the RawTemplate that can be marshalled to JSON.
func (r *RawTemplate) CreatePackerTemplate() (PackerTemplate, error) {
	var err error
	var vars map[string]interface{}

	Log.Info("Creating PackerTemplate from a RawTemplate.")

	// Set the src_dir and out_dir, in case there are variables embedded in them.
	// These can be embedded in other dynamic variables so they need to be resolved
	// first to avoid a mutation issue. Only Rancher static variables can be used
	// in these two Settings.
	// check src_dir and out_dir first:

	// Get the delim and set the replacement map
	r.varVals = map[string]string{r.delim + "type": r.Type, r.delim + "Release": r.Release, r.delim + "Arch": r.Arch, r.delim + "Image": r.Image, r.delim + "date": r.date }

	r.SrcDir = r.replaceVariables(r.SrcDir)
	r.OutDir = r.replaceVariables(r.OutDir)

	// Commands and scripts dir need to be resolved next
	r.varVals = map[string]string{r.delim + "out_dir": r.OutDir, r.delim + "src_dir": r.SrcDir, r.delim + "type": r.Type, r.delim + "Release": r.Release, r.delim + "Arch": r.Arch, r.delim + "Image": r.Image, r.delim + "date": r.date }

	r.CommandsDir = r.replaceVariables(r.CommandsDir)
	r.ScriptsDir = r.replaceVariables(r.ScriptsDir)

	// Create a full variable replacement map, know that the SrcDir and OutDir stuff are resolved.
	// Rest of the replacements are done by the packerers.
	r.varVals = map[string]string{r.delim + "out_dir": r.OutDir, r.delim + "src_dir": r.SrcDir, r.delim + "commands_dir": r.CommandsDir, r.delim + "scripts_dir": r.ScriptsDir, r.delim + "type": r.Type, r.delim + "Release": r.Release, r.delim + "Arch": r.Arch, r.delim + "Image": r.Image, r.delim + "date": r.date }
	
	// General Packer Stuff
	p := PackerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description

	// Builders
	iSl := make([]interface{}, len(r.Builders))

	if p.Builders, vars, err = r.createBuilders(); err != nil {
		Log.Crit(err.Error())
		fmt.Println(err.Error())
	}

	// Post-Processors
	var i int
	var sM map[string]interface{}
	iSl = make([]interface{}, len(r.PostProcessors))

	for k, pp := range r.PostProcessors {
		sM = pp.settingsToMap(k, r)

		if sM == nil {
			err = errors.New("an error occured while trying to create the Packer post-processor template for " + k)
			Log.Error(err.Error())
			return p, err
		}

		iSl[i] = sM
		i++
	}

	p.PostProcessors = iSl

	// Provisioners
	i = 0
	iM := make(map[string]interface{})
	iSl = make([]interface{}, len(r.PostProcessors))
	for k, pr := range r.Provisioners {
		iM = pr.settingsToMap(k, r)
		if iM == nil {
			err = fmt.Errorf("CreatePackerTemplate error: the Settings map is nil")
			Log.Crit(err.Error())
			return p, err
		}
	
		// If there are any scripts add them. Scripts are already in an array. Scripts use
		// a map[string]interface{} structure for JSON
		if len(pr.Scripts) > 0 {
			iM["scripts"] = pr.Scripts
		}
		iSl[i] = iM
		i++
	}

	p.Provisioners = iSl
	p.Variables = vars

	// Now we can create the Variable Section

	// Return the generated Packer Template
	Log.Info("PackerTemplate created from a RawTemplate.")
	return p, nil
}

func (r *RawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	// Takes a raw builder and create the appropriate Packer Builders along with a
	// slice of variables for that section builder type. Some Settings are in-lined
	// instead of adding them to the variable section.

	if r.BuilderType == nil {
		err = fmt.Errorf("RawTemplate.createBuilders() error: no builder types were passed")
		Log.Error(err.Error())
		return
	}
	fmt.Printf("createBuilders\n%+v", r.BuilderType)

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var k, val, v string
	var i, ndx int
	bldrs = make([]interface{}, len(r.BuilderType))
	for _, bType := range r.BuilderType {
		Log.Info("Creating builder from RawTemplate: " + bType)
		fmt.Printf("Creating builder for %+v\n", bType)
		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWare:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				fmt.Println(err.Error())				
				Log.Error(err.Error())
			}				
			tmpS["type"] = BuilderVMWare

			// Generate builder specific section
			tmpvm := make(map[string]string, len(r.Builders[bType].VMSettings))

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				val = r.replaceVariables(val)
				tmpvm[k] = val
			}
			tmpS["vmx_data"] = tmpvm

		case BuilderVBox:
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				Log.Error(err.Error())
				fmt.Println(err.Error())
				return
			}

			tmpS["type"] = BuilderVBox


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
			Log.Error(err.Error())
			return
		}
		fmt.Printf("ZZZZZZ%v:\n%+v\n%+v\n\n", bType, tmpS)
		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return
}

func (r *RawTemplate) replaceVariables(s string) string {
	// Checks incoMing string for variables and replaces them with their values.
	// Some of these are known and static to each build. For Rancher, this type
	// of variable is called a static variable. The other variables are dynamic-
	// ally resolved at build time via this methon. They may have their values 
	// change, during final resolution of the value for that variable, which can
	// be problematic when a variable references another, as this can leads to a
	// mutating variable issue issue. 
	//
	// This should only be an issue with `out_dir` and `src_dir` as it doesn't make
	// sense to embed any of the other non-static values within those directories. As
	// such, one of the restrictions on the `out_dir` and `src_dir` variables is that
	// they can only contain valid output information and static variables. 
	// 
	// All other variables can reference any of the static variables along with any
	// of the *_dir variables.
	//
	// Currently supported variable replacements:
	// 	STATIC				DYNAMIC
	//	Arch				commands_dir
	//	date (in ISO-8601 date format)	out_dir
	//	Image				scripts_dir
	//	type				src_dir

	//see if the delim is in the string
	if strings.Index(s, r.delim)  < 0 {
		return s
	}

	for vName, vVal := range r.varVals {
		s = strings.Replace(s, vName, vVal, -1)
	}

	return s
}

func (r *RawTemplate) variableSection() (map[string]interface{}, error) {
	// Generates the variable section.
	var v map[string]interface{}
	v = make(map[string]interface{})

	Log.Info("Creating variable section from RawTemplate")

	return v, nil
}

func (r *RawTemplate) commonVMSettings(old []string, new []string) (Settings map[string]interface{}, vars []string, err error) {
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
		
			var ret, notSupported string
			Settings[k] = v
			switch r.Type {
			case "ubuntu":
				rel := &ubuntu{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: strings.ToUpper(v)}, Arch: r.Arch, Distro: r.Type, Image: r.Image, Release: r.Release}}	
				ret = rel.SetISOFilename()

				// TODO wtf...not supported handling...
				if ret == "" {
					Log.Error("Error: retrieval of the checksum information for the requested " + rel.Distro + " ISO failed. No checksum values were retrieved")
					Settings["iso_checksum"] = notSupported
				} else {
					Settings["iso_checksum"] = ret
				}

				ret = rel.SetURL()
				if ret == "" {
					Log.Error("Error: retrieval of the URL information for the requested " + rel.Distro + " ISO failed. The ISO URL was not created.")
					Settings["iso_url"] = notSupported
				} else {
					Settings["iso_url"] = ret
				}	
			case "centos":
				notSupported = "Retrieval of CentOS ISO information has not been implemented."				
			default:
				notSupported = r.Type + " is not supported."
			}
	
			if notSupported != "" {
				Settings["iso_url"] = notSupported
				Settings["iso_checksum"] = ""
			} 

		case "boot_command", "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string	
			if commands, err = commandsFromFile(v); err != nil {
				Log.Error("error", err.Error())
				Settings[k] = "Error: " + err.Error()
			} else {
				// Boot commands are slices, the others are just a string.
				if k == "boot_command" {
					Settings[k] = commands
				} else {
					// Assume it's the first element.
					Settings[k] = commands[0]
				}
			}
			
		default:
			// just use the value
			Settings[k] = v
		}
		//		kname = getVariableName(k)

	}

	return
}

func (r *RawTemplate) mergeBuildSettings(bld RawTemplate) {
	// merges Settings between an old and new template. 
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's 
	// default template or from a defined build template.
	r.updateIODirInf(bld.IODirInf)
	r.updatePackerInf(bld.PackerInf)

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

func (r *RawTemplate) mergeDistroSettings(d distro) {
	// merges Settings between an old and new template. 
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's 
	// default template or from a defined build template.
	r.IODirInf.updateIODirInf(d.IODirInf)
	r.PackerInf.updatePackerInf(d.PackerInf)

	// If defined, Builders override any prior builder Settings
	if d.BuilderType != nil && len(d.BuilderType) > 0 {
		r.BuilderType = d.BuilderType
	} 
	
	// merge the build portions.
	r.Builders = getMergedBuilders(r.Builders, d.Builders)
	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, d.PostProcessors)
	r.Provisioners = getMergedProvisioners(r.Provisioners, d.Provisioners)

	return
}

func (io *IODirInf) updateIODirInf(new IODirInf) {
	if new.CommandsDir != "" {
		io.CommandsDir = new.CommandsDir
	}

	if new.OutDir != "" {
		io.OutDir = new.OutDir
	}
	
	if new.ScriptsDir != "" {
		io.ScriptsDir = new.ScriptsDir
	}

	if new.SrcDir != "" {
		io.SrcDir = new.SrcDir
	}

	return
}

func (pi PackerInf) updatePackerInf(new PackerInf) {
	if new.MinPackerVersion != "" {
		pi.MinPackerVersion = new.MinPackerVersion
	}

	if new.Description != "" {
		pi.Description = new.Description
	}
	return
}

func (b *build) updateBuild(new build) {
	return	
}

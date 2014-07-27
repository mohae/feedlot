package ranchr

import (
	"os"
	"strings"
	"time"

	json "github.com/mohae/customjson"
	jww "github.com/spf13/jwalterweatherman"
)

// rawTemplate holds all the information for a Rancher template. This is used
// to generate the Packer Build.
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

// mewRawTemplate returns a rawTemplate with current date in ISO 8601 format.
// This should be called when a rawTemplate with the current date is desired.
func newRawTemplate() rawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	r := rawTemplate{date: splitDate[0], delim: os.Getenv(EnvParamDelimStart)}
	return r
}

// r.createDistroTemplate assigns the default distro template values to a
// rawTemplate
func (r *rawTemplate) createDistroTemplate(d rawTemplate) {
	r.IODirInf = d.IODirInf
	r.PackerInf = d.PackerInf
	r.BuildInf = d.BuildInf
	r.Arch = d.Arch
	r.BaseURL = d.BaseURL
	r.Type = d.Type
	r.Image = d.Image
	r.Release = d.Release
	r.BuilderTypes = d.BuilderTypes
	r.Builders = d.Builders
	r.PostProcessorTypes = d.PostProcessorTypes
	r.PostProcessors = d.PostProcessors
//	r.ProvisionersType = d.ProvisionersType
//	r.Provisioners = d.Provisioners
	return
}

// r.createPackerTemplate creates a Packer template from the rawTemplate that
// can be marshalled to JSON.
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
	iSl = make([]interface{}, len(r.PostProcessors))

	if p.PostProcessors, vars, err = r.createPostProcessors(); err != nil {
		jww.ERROR.Println(err.Error())
		return p, err
	}

/*
	// Provisioners
//	i := 0
//	iM := make(map[string]interface{})
	iSl = make([]interface{}, len(r.Provisioners))

	if p.Provisioners, vars, err = r.createProvisioners(); err != nil {
		jww.ERROR.Println(err.Error())
		return p, err
	}
*/
/*	for k, pr := range r.Provisioners {
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

*/	p.Provisioners = iSl
	p.Variables = vars

	// Now we can create the Variable Section
	// TODO

	// Return the generated Packer Template
	jww.DEBUG.Println("PackerTemplate created from a rawTemplate.")

	return p, nil
}




// r.replaceVariables checks incoming string for variables and replaces them
// with their values.
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

// r.variableSection generates the variable section. This doesn't do anything
// at the moment.
func (r *rawTemplate) variableSection() (map[string]interface{}, error) {
	var v map[string]interface{}
	v = make(map[string]interface{})
	return v, nil
}

// r.mergeBuildSettings merges Settings between an old and new template. Note:
// Arch, Image, and Release are not updated here as how these fields are 
// updated depends on whether this is a build from a distribution's default 
// template or from a defined build template.
func (r *rawTemplate) mergeBuildSettings(bld rawTemplate) {
	jww.TRACE.Print(json.MarshalIndentToString(bld, "", indent))
	r.IODirInf.update(bld.IODirInf)
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)

	// If defined, Builders override any prior builder Settings.
	if bld.BuilderTypes != nil && len(bld.BuilderTypes) > 0 {
		r.BuilderTypes = bld.BuilderTypes
	}

	// merge the build portions.
//	r.Builders = getMergedBuilders(r.Builders, bld.Builders)
//	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, bld.PostProcessors)
//	r.Provisioners = getMergedProvisioners(r.Provisioners, bld.Provisioners)

	return
}

// r.mergeDistroSettings merges the settings...hmm the name doesn't
// seem to reflect what it actually is doing.
// TODO rename this 
func (r *rawTemplate) mergeDistroSettings(d *distro) {
	jww.TRACE.Printf("%v\n%v", json.MarshalIndentToString(r, "", indent), json.MarshalIndentToString(d, "", indent))
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)

	r.BuildInf.update(d.BuildInf)
	// If defined, Builders override any prior builder Settings
	if d.BuilderTypes != nil && len(d.BuilderTypes) > 0 {
		r.BuilderTypes = d.BuilderTypes
	}

	// merge the build portions.
	jww.TRACE.Printf("Merging old Builder: %v\nand new Builder: %v", json.MarshalIndentToString(r.Builders, "", indent), json.MarshalIndentToString(d.Builders, "", indent))
//	r.Builders = getMergedBuilders(r.Builders, d.Builders)
//	jww.TRACE.Printf("Merged Builder: %v", r.Builders)
//	r.PostProcessors = getMergedPostProcessors(r.PostProcessors, d.PostProcessors)
//	jww.TRACE.Printf("Merged PostProcessors: %v", r.PostProcessors)
//	r.Provisioners = getMergedProvisioners(r.Provisioners, d.Provisioners)
//	jww.TRACE.Printf("Merged Provisioners: %v", r.Provisioners)
	return
}

// Get a slice of script names from the shell provisioner, if any.
func (r *rawTemplate) ScriptNames() []string {
	jww.DEBUG.Println("Entering rawTemplate.ScriptNames...")
	var s []string
	//TODO
/*	if len(r.Provisioners["shell"].Scripts) > 0 {
		s = make([]string, len(r.Provisioners["shell"].Scripts))

		for i, script := range r.Provisioners["shell"].Scripts {
			//explode on "/"
			parts := strings.Split(script, "/")

			// the last element is the script name
			s[i] = parts[len(parts)-1]
		}

	}
*/
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

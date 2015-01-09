package ranchr

import (
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mohae/deepcopy"
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
	// Current date in ISO 8601
	date string
	// The character(s) used to identify variables for Rancher. By default
	// this is a colon, :. Currently only a starting delimeter is supported.
	delim string
	// The distro that this template targets. The type must be a supported
	// type, i.e. defined in supported.toml. The values for type are
	// consistent with Packer values.
	Distro string
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
func newRawTemplate() *rawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	return &rawTemplate{date: splitDate[0], delim: os.Getenv(EnvParamDelimStart)}
}

// r.createPackerTemplate creates a Packer template from the rawTemplate that
// can be marshalled to JSON.
func (r *rawTemplate) createPackerTemplate() (packerTemplate, error) {
	var err error
	// Resolve the Rancher variables to their final values.
	r.mergeVariables()
	// General Packer Stuff
	p := packerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description
	// Builders
	p.Builders, _, err = r.createBuilders()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Post-Processors
	p.PostProcessors, _, err = r.createPostProcessors()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Provisioners
	p.Provisioners, _, err = r.createProvisioners()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Now we can create the Variable Section
	// TODO: currently not implemented/supported
	// Return the generated Packer Template
	return p, nil
}

// replaceVariables checks incoming string for variables and replaces them
// with their values.
func (r *rawTemplate) replaceVariables(s string) string {
	//see if the delim is in the string, if not, nothing to replace
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

// r.setDefaults takes the incoming distro settings and merges them with its
// existing settings, which are set to rancher's defaults, to create the
// default template.
func (r *rawTemplate) setDefaults(d *distro) {
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)
	r.BuildInf.update(d.BuildInf)
	// If defined, BuilderTypes override any prior BuilderTypes Settings
	if d.BuilderTypes != nil && len(d.BuilderTypes) > 0 {
		r.BuilderTypes = d.BuilderTypes
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if d.PostProcessorTypes != nil && len(d.PostProcessorTypes) > 0 {
		r.PostProcessorTypes = d.PostProcessorTypes
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if d.ProvisionerTypes != nil && len(d.ProvisionerTypes) > 0 {
		r.ProvisionerTypes = d.ProvisionerTypes
	}
	// merge the build portions.
	r.updateBuilders(d.Builders)
	r.updatePostProcessors(d.PostProcessors)
	r.updateProvisioners(d.Provisioners)
	return
}

// r.updateBuildSettings merges Settings between an old and new template. Note:
// Arch, Image, and Release are not updated here as how these fields are
// updated depends on whether this is a build from a distribution's default
// template or from a defined build template.
func (r *rawTemplate) updateBuildSettings(bld *rawTemplate) {
	r.IODirInf.update(bld.IODirInf)
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)
	// If defined, Builders override any prior builder Settings.
	if bld.BuilderTypes != nil && len(bld.BuilderTypes) > 0 {
		r.BuilderTypes = bld.BuilderTypes
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if bld.PostProcessorTypes != nil && len(bld.PostProcessorTypes) > 0 {
		r.PostProcessorTypes = bld.PostProcessorTypes
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if bld.ProvisionerTypes != nil && len(bld.ProvisionerTypes) > 0 {
		r.ProvisionerTypes = bld.ProvisionerTypes
	}
	// merge the build portions.
	r.updateBuilders(bld.Builders)
	r.updatePostProcessors(bld.PostProcessors)
	r.updateProvisioners(bld.Provisioners)
}

// Get a slice of script names from the shell provisioner, if any.
func (r *rawTemplate) ScriptNames() []string {
	scripts := "scripts"
	// See if there is a shell provisioner
	_, ok := r.Provisioners[Shell.String()]
	if !ok {
		return nil
	}
	// See if there shell provisioner array section contains scripts
	_, ok = r.Provisioners[Shell.String()].Arrays[scripts]
	if !ok {
		return nil
	}
	scrpts := deepcopy.InterfaceToSliceStrings(r.Provisioners[Shell.String()].Arrays[scripts])
	names := make([]string, len(scrpts))
	for i, scrpt := range scrpts {
		so := reflect.ValueOf(scrpt)
		parts := strings.Split(so.Interface().(string), "/")
		// the last element is the script name
		names[i] = parts[len(parts)-1]
	}
	return names
}

// Set the src_dir and out_dir, in case there are variables embedded in them.
// These can be embedded in other dynamic variables so they need to be resolved
// first to avoid a mutation issue. Only Rancher static variables can be used
// in these two Settings.
func (r *rawTemplate) mergeVariables() {
	// check src_dir and out_dir first:
	// Get the delim and set the replacement map, resolve name information
	r.varVals = map[string]string{r.delim + "distro": r.Distro, r.delim + "release": r.Release, r.delim + "arch": r.Arch, r.delim + "image": r.Image, r.delim + "date": r.date, r.delim + "build_name": r.BuildName}
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

// ISOInfo sets the ISO info for the template's supported distro type. This
// also sets the builder specific string, when applicable.
func (r *rawTemplate) ISOInfo(builderType Builder, settings []string) error {
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
	switch r.Distro {
	case "ubuntu":
		r.releaseISO = &ubuntu{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: checksumType}, Arch: r.Arch, Distro: r.Distro, Image: r.Image, Release: r.Release}}
		r.releaseISO.SetISOInfo()
		r.osType, err = r.releaseISO.(*ubuntu).getOSType(builderType.String())
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	case "centos":
		r.releaseISO = &centOS{release: release{iso: iso{BaseURL: r.BaseURL, ChecksumType: checksumType}, Arch: r.Arch, Distro: r.Distro, Image: r.Image, Release: r.Release}}
		r.releaseISO.SetISOInfo()
		r.osType, err = r.releaseISO.(*centOS).getOSType(builderType.String())
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	return nil
}

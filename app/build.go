package app

import (
	"fmt"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	contour.RegisterBoolFlag("example", "eg", false, "false", "whether this is an example")
	contour.RegisterStringFlag("exampledir", "ed", "examples/", "examples/", "example directory")
	contour.RegisterBoolFlag("optionals", "", false, "false", "include optional flags")
	contour.RegisterStringFlag("builders", "", "", "", "override build types")
	contour.RegisterStringFlag("postprocessors", "", "", "", "override postprocessor types")
	contour.RegisterStringFlag("provisioners", "", "", "", "override provisioner types")
}

// BuildDistro creates a build based on the target distro's defaults. The
// ArgsFilter contains information on the target distro and any overrides that
// are to be applied to the build.
//
// Returns an error or nil if successful.
func BuildDistro(a ArgsFilter) error {
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("BuildDistro failed: %s", err.Error())
			jww.ERROR.Println(err)
			return err
		}
	}
	err := buildPackerTemplateFromDistro(a)
	if err != nil {
		err = fmt.Errorf("BuildDistro failed: %s", err.Error())
		jww.ERROR.Println(err)
		return err
	}
	// TODO: what does this argString processing do, or supposed to do? and document it this time!
	argString := ""
	if a.Arch != "" {
		argString += "Arch=" + a.Arch
	}
	if a.Image != "" {
		if argString != "" {
			argString += ", "
		}
		argString += "Image=" + a.Image
	}
	if a.Release != "" {
		if argString != "" {
			argString += ", "
		}
		argString += "Release=" + a.Release
	}
	return nil

}

// Create Packer templates from specified build templates.
func buildPackerTemplateFromDistro(a ArgsFilter) error {
	var rTpl rawTemplate
	if a.Distro == "" {
		err := fmt.Errorf("cannot build a Packer template because no there wasn't a value for the distro flag")
		jww.ERROR.Println(err)
		return err
	}
	// Get the default for this distro, if one isn't found then it isn't Supported.
	rTpl, err := DistroDefaults.GetTemplate(a.Distro)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// If any overrides were passed, set them.
	if a.Arch != "" {
		rTpl.Arch = a.Arch
	}
	if a.Image != "" {
		rTpl.Image = a.Image
	}
	if a.Release != "" {
		rTpl.Release = a.Release
	}
	rTpl.BuildName = ":type-:release-:arch-:image-rancher"

	//	// make a copy of the .
	//	rTpl := newRawTemplate()
	//	rTpl.updateBuilders(d.Builders)

	// Since distro builds don't actually have a build name, we create one
	// out of the args used to create it.
	// TODO: given the above, should this be done? Or should the buildname for distro
	//       builds be merged later?
	rTpl.BuildName = fmt.Sprintf("%s-%s-%s-%s", rTpl.Distro, rTpl.Release, rTpl.Arch, rTpl.Image)
	pTpl := packerTemplate{}
	// Now that the raw template has been made, create a Packer template out of it
	pTpl, err = rTpl.createPackerTemplate()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// Create the JSON version of the Packer template. This also handles creation of
	// the build directory and copying all files that the Packer template needs to the
	// build directory.
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.dirs, rTpl.files)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

// BuildBuilds manages the process of creating Packer Build templates out of
// the passed build names. All builds are done concurrently.  Returns either a
// message providing information about the processing of the requested builds
// or an error.
func BuildBuilds(buildNames ...string) (string, error) {
	if buildNames[0] == "" {
		err := fmt.Errorf("builds failed: no build name was received")
		jww.ERROR.Println(err)
		return "", err
	}
	// Only load supported if it hasn't been loaded. Even though LoadSupported
	// uses a mutex to control access to prevent race conditions, no need to
	// call it if its already loaded.
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("builds failed: %s", err.Error())
			jww.ERROR.Println(err)
			return "", err
		}
	}
	// First load the build information
	err := loadBuilds()
	if err != nil {
		err = fmt.Errorf("builds failed: %s", err)
		jww.ERROR.Println(err)
		return "", err
	}
	// Make as many channels as there are build requests.
	var errorCount, builtCount int
	nBuilds := len(buildNames)
	doneCh := make(chan error, nBuilds)
	// Process each build request
	for i := 0; i < nBuilds; i++ {
		go buildPackerTemplateFromNamedBuild(buildNames[i], doneCh)
	}
	// Wait for channel done responses.
	for i := 0; i < nBuilds; i++ {
		err := <-doneCh
		if err != nil {
			jww.ERROR.Println(err)
			errorCount++
		} else {
			builtCount++
		}
	}
	if nBuilds == 1 {
		if builtCount > 0 {
			return fmt.Sprintf("%s was successfully processed and its Packer template was created", buildNames[0]), nil
		}
		return fmt.Sprintf("Processing of the %s build failed with an error.", buildNames[0]), nil
	}
	return fmt.Sprintf("BuildBuilds: %v Builds were successfully processed and their Packer templates were created, %v Builds were unsucessfully process and resulted in errors..", builtCount, errorCount), nil
}

// buildPackerTemplateFromNamedBuild creates a Packer tmeplate and associated
// artifacts for the passed build.
func buildPackerTemplateFromNamedBuild(name string, doneCh chan error) {
	if name == "" {
		err := fmt.Errorf("unable to build Packer template: no build name was received")
		doneCh <- err
		return
	}
	var ok bool
	// Check the type and create the defaults for that type, if it doesn't already exist.
	rTpl := rawTemplate{}
	bTpl, err := getBuildTemplate(name)
	if err != nil {
		doneCh <- fmt.Errorf("processing of build template %q failed: %s", name, err.Error())
		return
	}
	// See if the distro default exists.
	rTpl, ok = DistroDefaults.Templates[DistroFromString(bTpl.Distro)]
	if !ok {
		doneCh <- fmt.Errorf("building Packer template for %s failed: an unsupported distro, %s, was specified", name, bTpl.Distro)
		return
	}
	// Set build iso information overrides, if any.
	if bTpl.Arch != "" {
		rTpl.Arch = bTpl.Arch
	}
	if bTpl.Image != "" {
		rTpl.Image = bTpl.Image
	}
	if bTpl.Release != "" {
		rTpl.Release = bTpl.Release
	}
	bTpl.Name = name
	// create build template() then call create packertemplate
	rTpl.build = DistroDefaults.Templates[DistroFromString(bTpl.Distro)].build
	rTpl.updateBuildSettings(bTpl)
	pTpl := packerTemplate{}
	pTpl, err = rTpl.createPackerTemplate()
	if err != nil {
		doneCh <- err
		return
	}
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.dirs, rTpl.files)
	if err != nil {
		doneCh <- err
		return
	}
	doneCh <- nil
	return
}

func getBuildTemplate(name string) (*rawTemplate, error) {
	return nil, nil
}

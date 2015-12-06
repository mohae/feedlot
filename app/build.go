package app

import (
	"fmt"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

// BuildDistro creates a build based on the target distro's defaults. The
// ArgsFilter contains information on the target distro and any overrides that
// are to be applied to the build.  Returns either a processing message or an
// error.
func BuildDistro() (string, error) {
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("BuildDistro failed: %s", err)
			jww.ERROR.Println(err)
			return "", err
		}
	}
	message, err := buildPackerTemplateFromDistro()
	if err != nil {
		err = fmt.Errorf("BuildDistro failed: %s", err)
		jww.ERROR.Println(err)
	}
	return message, err

}

// Create Packer templates from specified build templates.
// TODO: refactor to match updated handling
func buildPackerTemplateFromDistro() (string, error) {
	var rTpl *rawTemplate
	// Get the default for this distro, if one isn't found then it isn't
	// Supported.
	rTpl, err := DistroDefaults.GetTemplate(contour.GetString("distro"))
	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	// If there were any overrides, set them.
	if contour.GetString("arch") != "" {
		rTpl.Arch = contour.GetString("arch")
	}
	if contour.GetString("image") != "" {
		rTpl.Image = contour.GetString("image")
	}
	if contour.GetString("release") != "" {
		rTpl.Release = contour.GetString("release")
	}

	// Since distro builds don't actually have a build name, we create one out
	// of the args used to create it.
	rTpl.BuildName = fmt.Sprintf("%s-%s-%s-%s", rTpl.Distro, rTpl.Release, rTpl.Arch, rTpl.Image)
	pTpl := packerTemplate{}
	// Now that the raw template has been made, create a Packer template out of it
	pTpl, err = rTpl.createPackerTemplate()
	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	// Create the JSON version of the Packer template. This also handles
	// creation of the build directory and copying all files that the Packer
	// template needs to the build directory.
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.dirs, rTpl.files)
	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}
	return fmt.Sprintf("build for %q complete: Packer template name is %q", rTpl.Distro, rTpl.BuildName), nil
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
	// Only load supported if it hasn't been loaded.
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("builds failed: %s", err)
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
	// Make as many channels as there are build requests.  A channel per build
	// is fine for now.  If a large number of builds needs to be supported,
	// switching to a queue and worker pool would be a better choice.
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
	bTpl, err := getBuildTemplate(name)
	if err != nil {
		doneCh <- fmt.Errorf("processing of build template %q failed: %s", name, err)
		return
	}
	// See if the distro default exists.
	rTpl := rawTemplate{}
	rTpl, ok = DistroDefaults.Templates[DistroFromString(bTpl.Distro)]
	if !ok {
		doneCh <- fmt.Errorf("creation of Packer template for %s failed: %s not supported", name, bTpl.Distro)
		return
	}
	// TODO: this is probably where the merging of parent build would occur
	rTpl.Name = name
	rTpl.updateBuildSettings(bTpl)
	if contour.GetBool(Example) {
		rTpl.IsExample = true
		rTpl.ExampleDir = contour.GetString(ExampleDir)
		rTpl.setExampleDirs()
	}
	pTpl, err := rTpl.createPackerTemplate()
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

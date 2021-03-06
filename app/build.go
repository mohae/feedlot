package app

import (
	"fmt"

	"github.com/mohae/contour"
	"github.com/mohae/feedlot/conf"
	"github.com/mohae/feedlot/log"
)

// BuildDistro creates a build based on the target distro's defaults. The
// ArgsFilter contains information on the target distro and any overrides that
// are to be applied to the build.  Returns either a processing message or an
// error.
func BuildDistro() (string, error) {
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("build packer template from distro failed: %s", err)
			log.Error(err)
			return "", err
		}
		log.Debug("build distro: set distro defaults")
	}
	message, err := buildPackerTemplateFromDistro()
	if err != nil {
		err = fmt.Errorf("build packer template from distro failed: %s", err)
		log.Error(err)
	} else {
		log.Infof("build distro: %s", message)
	}
	return message, err
}

// Create Packer templates from specified build templates.
// TODO: refactor to match updated handling
func buildPackerTemplateFromDistro() (string, error) {
	var rTpl *RawTemplate
	d := contour.GetString("distro")
	log.Infof("%s: create template using distro defaults", d)
	// Get the default for this distro, if one isn't found then it isn't
	// Supported.
	rTpl, err := DistroDefaults.GetTemplate(d)
	if err != nil {
		err = Error{slug: "get template", err: err}
		log.Errorf("%s: %s", d, err)
		return "", err
	}
	log.Debugf("%s: got distro defaults", d)
	// If there were any overrides, set them.
	if contour.GetString("arch") != "" {
		rTpl.Arch = contour.GetString("arch")
		log.Debugf("%s: set template arch to %s", d, rTpl.Arch)
	}
	if contour.GetString("image") != "" {
		rTpl.Image = contour.GetString("image")
		log.Debugf("%s: set template arch to %s", d, rTpl.Image)
	}
	if contour.GetString("release") != "" {
		rTpl.Release = contour.GetString("release")
		log.Debugf("%s: set template arch to %s", d, rTpl.Release)
	}

	// Since distro builds don't actually have a build name, we create one out
	// of the args used to create it.
	rTpl.BuildName = fmt.Sprintf("%s-%s-%s-%s", rTpl.Distro, rTpl.Release, rTpl.Arch, rTpl.Image)
	log.Infof("%s: build name: %s", d, rTpl.BuildName)
	// Now that the raw template has been made, create a Packer template out of it
	pTpl, err := rTpl.createPackerTemplate()
	if err != nil {
		err = Error{slug: "get template", err: err}
		log.Errorf("%s: %s", d, err)
		return "", err
	}
	// Create the JSON version of the Packer template. This also handles
	// creation of the build directory and copying all files that the Packer
	// template needs to the build directory.
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.Dirs, rTpl.Files)
	if err != nil {
		log.Errorf("%s: %s", d, err)
		return "", err
	}
	msg := fmt.Sprintf("%s: build complete: Packer template name is %q", rTpl.Distro, rTpl.BuildName)
	log.Info(msg)
	return msg, nil
}

// BuildBuilds manages the process of creating Packer Build templates out of
// the passed build names. All builds are done concurrently.  Returns either a
// message providing information about the processing of the requested builds
// or an error.
func BuildBuilds(buildNames ...string) (string, error) {
	if buildNames[0] == "" {
		err := fmt.Errorf("build builds failed: no build names were received")
		log.Error(err)
		return "", err
	}
	log.Infof("start builds for: %v", buildNames)
	// Only load supported if it hasn't been loaded.
	if !DistroDefaults.IsSet {
		log.Debug("loading distro defaults")
		err := DistroDefaults.Set()
		if err != nil {
			err = fmt.Errorf("builds failed: %s", err)
			log.Error(err)
			return "", err
		}
	}
	// First load the build information
	err := loadBuilds()
	if err != nil {
		err = fmt.Errorf("builds failed: %s", err)
		log.Error(err)
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
		log.Debugf("%s: start build", buildNames[i])
		go buildPackerTemplateFromNamedBuild(buildNames[i], doneCh)
	}
	// Wait for channel done responses.
	for i := 0; i < nBuilds; i++ {
		err := <-doneCh
		if err != nil {
			log.Error(err)
			errorCount++
		} else {
			log.Debug("a template was successfully created")
			builtCount++
		}
	}
	var msg string
	if nBuilds == 1 {
		if builtCount > 0 {
			msg = fmt.Sprintf("%s was successfully processed and its packer template was created", buildNames[0])
			goto done
		}
		msg = fmt.Sprintf("%s: build failed with an error.", buildNames[0])
		goto done
	}
	msg = fmt.Sprintf("build builds: %v builds were successfully processed and their packer templates were created, %v builds were unsucessfully process and resulted in errors..", builtCount, errorCount)
done:
	log.Info(msg)
	return msg, nil
}

// buildPackerTemplateFromNamedBuild creates a Packer tmeplate and associated
// artifacts for the passed build.
func buildPackerTemplateFromNamedBuild(name string, doneCh chan error) {
	if name == "" {
		err := fmt.Errorf("build packer template failed: no build name was received")
		doneCh <- err
		return
	}
	log.Infof("%s: start creation of packer template", name)
	defer log.Infof("%s: end creation of packer template", name)
	var ok bool
	// Check the type and create the defaults for that type, if it doesn't already exist.
	bTpl, err := getBuildTemplate(name)
	if err != nil {
		doneCh <- fmt.Errorf("%s: build failed: %s", name, err)
		return
	}
	// See if the distro default exists.
	rTpl := RawTemplate{}
	rTpl, ok = DistroDefaults.Templates[ParseDistro(bTpl.Distro)]
	if !ok {
		err := fmt.Errorf("%s: %s: not a supported distro", name, bTpl.Distro)
		log.Error(err)
		doneCh <- err
		return
	}
	// TODO: this is probably where the merging of parent build would occur
	rTpl.Name = name
	err = rTpl.updateBuildSettings(bTpl)
	if err != nil {
		err = Error{name, err}
		log.Error(err)
		doneCh <- err
	}
	if contour.GetBool(conf.Example) {
		log.Debugf("%s: using examples", name)
		rTpl.IsExample = true
		rTpl.ExampleDir = contour.GetString(conf.ExampleDir)
		rTpl.setExampleDirs()
	}
	pTpl, err := rTpl.createPackerTemplate()
	if err != nil {
		err = Error{name, err}
		log.Error(err)
		doneCh <- err
		return
	}
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.Dirs, rTpl.Files)
	if err != nil {
		err = Error{name, err}
		log.Error(err)
		doneCh <- err
		return
	}
	doneCh <- nil
	return
}

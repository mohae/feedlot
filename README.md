rancher
=======
Ranchers supply products to packers. Rancher creates templates Packer templates. Ok, a bit of a stretch, but it's the best I could come up with.

I created this because I found Packer templates more painful to work with than I would like, and why should I have to figure out the ISO information? Also there was a lot of replication between builders. So I did what any sane programmer would do, I wrote a program to do it for me. Because, that's a totally easier and simpler to do...

As it stands, this is a mvp with support for only VirtualBox and VMWare buiilders, the shell provisioner, and vagrant post-processor. Overrides are not supported for provisioners and post-processors. 

The Packer templates are generated from either the default Packer settings for a supported distro or from a Rancher build template. Rancher saves the results of a build to an directory of the same name within the output directory. This includes the .json file, referenced script files, and a configuration file for unattended installs of the distribution.

If the output directory for a given build already contains artifacts, Rancher will archive the target directory and save it as a .tgz using the directory name and the current date and time in a slightly modified ISO 8601 format, the `:` are stripped from the time in the filename. The old artifacts will then be deleted so that Rancher can ensure that the output from the current build will have a clean directory to write to.

Supported Distros:
    * ubuntu
    * centos

##Running Rancher##
Build a Packer template for CentOS using the distro defaults:

	rancher build -distro=centos

Build a Packer template for Ubuntu using the distro defaults with an image override:

	rancher build -distro=ubuntu -image=desktop

Build a Packer template from a named build:

	rancher build 1404-go

Build a Packer template for Ubuntu using the distro defaults and from more than one named build:
	
	rancher build -distro=centos -arch=i386 1404-go 6-lamp

##`rancher.cfg` and Environment Variables##
The `rancher.cfg` file is the default core configuration file for Rancher. It contains the default locations for all of the TOML files that Rancher uses. Environment variables are supported. Rancher will first check to see if the environment variable is set. If it is empty, the relevant `rancher.cfg` setting will be used.

For a full list of environment variables, please check the docs.

##Rancher Configuration files##
Information about defaults, supported distros, builds, and build lists are all stored in rancher configuration files, which are written in TOML. The defaults are set so that one can create a basic Packer template without any additional changes or overrides. 

For custom builds, only the differences between the default configuration and your desired configuration need to be set. Any missing values will use the default configuration for that build's distribution. These custom build configurations are stored in the rancher/conf.d/ directory.

If the defaults shipped with rancher are not to your liking, the default settings files, defaults.toml and supported.toml, can be modified. These files are found in the rancher/conf/ directory. Care should be taken when modifiying these settings.

The configuration files are described in order of precedence with later declarations overriding the prior ones.

###`conf/defaults.toml`###
The `defaults.toml` file contain most of the defaults for builds. A few settings that only apply at the distro level, like `base_url`, are not set here. The values in this file will be used, unless they are overridden later.

###`conf/supported.toml`###
The `supported.toml` file defines the supported operating systems, referred to as distros or distributions, and their settings, including any distro specific overrides. Because Rancher does distro specific processing, like ISO information look-up, Packer templates can be generated for supported distributions only. 

Each supported distro has a `defaults` section, which defines the default `architecture`, `image`, and `release` for the distro. These defaults define the target ISO to use.

####Distro Defaults####
The settings in the `defaults.toml` and `supported.toml` files are merged and the result, for each supported distro, being the distribution's default template. This is used for Rancher builds that use thee `-distro` flag. They are also used as the basis for the templates defined in `builds.toml`. 

####CentOS specific####
For CentOS, the `base_url` should only be used if you want to use a specific iso redirect mirror. The mirror url should be the url download string, less the iso name. So an ISO located at `http://bay.uchicago.edu/centos/6.5/isos/x86_64/CentOS-6.5-c86_64-minimal.iso` would have a `base_url` of `http://bay.uchicago.edu/centos/6.5/isos/x86_64/`.

If the `base_url` is left empty, the url will be randomly selected from that isoredirect.centos.org page for the desired CentOS version and architecture.

###`conf.d/builds.toml`###
`builds.toml` contains all custom builds for Rancher. Each build has a name, which also becomes the name of the Packer template json file and the output directory for the resulting Packer template and its resources. Build specific settings and overrides are set here.

Each build, at minimum, must have a `type`, which is named after Packer's `type` and represents the target distro.

###`conf.d/buildlists.toml`###
`buildlists.toml` defines named lists of builds. These name lists are used in conjunction with the `rancher buildlists` command.

##Rancher Commands##
Rancher uses mitchellh's cli package, with support for the `build` and `buildlists` custom commands, along with the standard `version` and `help` commands.

###`build`###
`rancher build buildNames...`

Supported Flags:

* -distro=<distroname>
* -arch=<arch>
* -image=<image>
* -release=<release>

If the `-distro` flag is passed, a build based on the default setting for the distro will be created. The additional flags allow for runtime overrides of the distro defaults for the target ISO. This flag can be used in conjunction with named builds. If both the -distro flag is passed along with a space separated list of one or more named builds are passed to the `build` sub-command, both the default Packer template for the distro and all of the Packer templates for the passed build names will be created.

###`buildlists`###
`rancher buildlists buildlistName...`

Generate the Packer templates for all of the builds listed within the passed buildlists. No flags are supported.

##Issues##
 * The buildlist command is not supported.
 * Some comments are outdated and need to be revised accordingly
 * preseed.cfg not tested (some may just be touch artifacts)
 * ks.cfg not tested (some may just be touch artifacts)
 * commands not tested (some may just be touch artifacts)
 * scripts not tested (some may just be touch artifacts)

##Other notes:##
 * Fully commented example toml files not created.
 * Docs needed
 * For CentOS, the ISO resolution is done on a per builder basis, which means that a Packer template could point to different ISO locations. Caching the first results will probably be added so that the locations are consistently used.




rancher
=======

>I am rarely happier than when spending entire day programming my computer to perform automatically a task that it would otherwise take me a good ten seconds to do by hand
> 
>   -Douglas Adams, _Last chance to See_

Ranchers supply products to packers to process and package. Rancher creates Packer templates for Packer to process and generate artifacts. Ok, a bit of a stretch, but it's the best I could come up with.

Rancher has default settings for Packer template generation. Each supported distribution has its own settings which are applied to created the distribution's defaultu template settings. 

Custom Packer templates can be specified via Rancher builds. A build is a named specification for a Packer template. Rancher saves the results of a build to a directory of the same name within Rancher's output directory. This includes the .json file, referenced script files, and a configuration file for unattended installs of the distribution.

If the output directory for a given build already contains artifacts, Rancher will archive the target directory and save it as a .tgz using the directory name and the current date and time in a slightly modified ISO 8601 format, the `:` are stripped from the time in the filename. The old artifacts will then be deleted so that Rancher can ensure that the output from the current build will have a clean directory to write to.

_There may be bugs in any of the following. The ones with __needs testing__ definitely need to be verified prior to using. Please report any  issues_

### Supported Distro
    * ubuntu
    * centos

### Supported Builders
    * vmware-iso
    * vmware-vmx __needs testing__
    * virtualbox-iso
    * virtualbox-ovf __needs testing__

### Supported Post-processors
    * vagrant
    * vagrant cloud

### Supported Provisioners 
    * ansible-local __needs testing__
    * file
    * salt-masterless __needs testing__
    * shell

##Running Rancher##
Build a Packer template for CentOS using the distro defaults:

	rancher build -distro=centos

Build a Packer template for Ubuntu using the distro defaults with an image override:

	rancher build -distro=ubuntu -image=desktop

Build a Packer template from a named build:

	rancher build centos6

Build a Packer template for Ubuntu using the distro defaults and from more than one named build:
	
	rancher build -distro=centos -arch=i386 1404-dev centos6

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

Each supported distro has a `defaults` section, which defines the default `architecture`, `image`, and `release` for the distro. These defaults define the target ISO to use unless the defaults are overridden by either flags or configuration.

####Distro Defaults####
The settings in the `defaults.toml` and `supported.toml` files are merged and the result, for each supported distro, being the distribution's default template. This is used for Rancher builds that use thee `-distro` flag. They are also used as the basis for the templates defined in `builds.toml`. 

####CentOS specific####
For CentOS, the `base_url` should only be used if you want to use a specific iso redirect mirror. The mirror url should be the url download string, less the iso name. So an ISO located at `http://bay.uchicago.edu/centos/6.5/isos/x86_64/CentOS-6.5-c86_64-minimal.iso` would have a `base_url` of `http://bay.uchicago.edu/centos/6.5/isos/x86_64/`. No validation is done on the `base_url` so you need to make sure that the URL is correct for your release and architecture.

If the `base_url` is left empty, the url will be randomly selected from that isoredirect.centos.org page for the desired CentOS version and architecture.

###`conf.d/builds.toml`###
`builds.toml` contains all custom builds for Rancher. Each build has a name, which also becomes the name of the Packer template json file and the output directory for the resulting Packer template and its resources. Build specific settings and overrides are set here.

Each build, at minimum, must have a `type`, which is named after Packer's `type` and represents the target distro.

Builds are used in conjunction with the `build` command or are added to `buildlists.toml` as part of a buildlist.

###`conf.d/buildlists.toml`###
`buildlists.toml` defines named lists of builds. These name lists are used in conjunction with the `rancher buildlists buildlistNames...` command. If you often find that the `build` sub-command is used with mupltiple build names, a buildlist might be useful instead. 

##Rancher Commands##
Rancher uses mitchellh's cli package, with support for the `build` and `run` custom commands, along with the standard `version` and `help` commands.

###`build`###
`rancher build <flags> buildNames...`

Supported Flags:

* -distro=<distro name>
* -arch=<architecture>
* -image=<image>
* -release=<release>

If the `-distro` flag is passed, a build based on the default setting for the distro will be created. The additional flags allow for runtime overrides of the distro defaults for the target ISO. This flag can be used in conjunction with named builds. If both the -distro flag is passed along with a space separated list of one or more named builds are passed to the `build` sub-command, both the default Packer template for the distro and all of the Packer templates for the passed build names will be created.

###`run`###
`rancher run buildlist...`

Generate the Packer templates for all of the builds listed within the passed buildlists; buildlists is variadic. No flags are supported.

##Notes:##
 * Fully commented example toml files not created.
 * Docs needed
 * Check Wiki for more info.

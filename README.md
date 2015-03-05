rancher
=======
>I am rarely happier than when spending entire day programming my computer to perform automatically a task that it would otherwise take me a good ten seconds to do by hand
> 
>   -Douglas Adams, _Last chance to See_

Ranchers supply products to packers to process and package. Rancher creates Packer templates for Packer to process and generate artifacts. Ok, a bit of a stretch, but it's the best I could come up with and better than the other names I had thought of.

Rancher has default settings for Packer template generation, `defaults.toml`. Each supported distribution has its own settings, found in `supported.toml` which are applied to created the distribution's default template settings. Both of these files are in the `conf/` directory.

Custom Packer templates can be specified via Rancher builds, `builds.toml`. A build is a named specification for a Packer template. Rancher saves the results of a build to a directory of the same name within Rancher's output directory. This includes the .json file, referenced script files, and any other assets required for the Packer build. It does not include the `iso` file, if applicable, or any other referenced resources.

If the output directory for a given build already contains artifacts, Rancher will archive the target directory and save it as a .tgz using the directory name and the current date and time in a slightly modified ISO 8601 format--the `:` are stripped from the time in the filename. The old artifacts will then be deleted so that Rancher can ensure that the output from the current build will have a clean directory to write to.

## Why rancher
I originally started Rancher because I didn't like the looking up the iso and checksum information for my builds. I also wasn't happy about replicating setting values between builders as I was too likely to make a mistake. I naively thought generating Packer templates from pre-defined configurations would be easier. During which, I realized I'd have to support more than just builders, so Rancher evolved into generating the complete template. 

The goal was to make it easier to create various configurations by minimally defining the differences between the named build and the distro's defaults.

I chose TOML as the configuration file type, but there's no reason why YAML or something else couldn't be supported. 

## Gotcha
There is one problem with my approach, parts of the Packer template depend on order for execution, which isn't supported with Go's JSON, as maps are not ordered by definition. Now, this is techinically Packer not following the JSON standard, but it works when the JSON files are manually created. 

As such, for now, if the order is important, you will have to manually re-order the sections within the generated JSON file. I have ideas on how to resolve this, which would involve using the order of the entries in the `*_type` arrays, outputting each JSON section separately to a temp file, and concatonating those files in the order defined by the `*_type` arrays. However that is not something I have done anything more than think about.

This will not be an issue for most templates.

## Supported Packer Section types
Each supported Packer section type has various configuration options. Any required configuration option is supported and a missing required element will result in a processing error, with the information logged to the Log. Most optional configuration options are supported, any that are not will be listed under as a `not supported configuration option` in that section's comments. If there are any options defined in the template that either do not exist for or are not supported by that Packer section type, a warning message will be logged, but the processing of the template will continue.

### Supported Distro
    * centos
    * debian
    * ubuntu


### Supported Builders
    * amazon-ebs
    * digitalocean
    * docker
    * googlecompute
    * virtualbox-iso
    * virtualbox-ovf
    * vmware-iso
    * vmware-vmx
    
### Supported Post-processors
    * compress
    * docker-import
    * docker-push
    * docker-save
    * docker-tag
    * vagrant
    * vagrant-cloud
    * vsphere

### Supported Provisioners 
    * ansible-local
    * file-uploads
    * salt-masterless
    * shell-scripts

## Rancher variable replacement
Rancher supports a limited number of variables in the toml configuration files. These are mostly used to allow for the path of files and names of things to be built based on other information within the configuration, e.g :distro is replaced by the name of the distro for which the Packer template is being built.

Since Packer uses Go's template engine, Rancher variables are prefixed with a colon, `:`, to avoid collision with Go template conventions, `{{}}`. Using a `:` in non-variable values may lead to unexpected results, as such don't use it unless you are prefixing a variable. Please submit an issue if this is problematic for your use-case scenario.

Current Rancher variable replacement is dumb, it will look for the variables that it can replace and do so. A setting string may have contain multiple variables, in which case it replaces them as encountered, with the exception of `src_dir` and `out_dir`, which are resolved before any other directory path variables. This is because most path settings will start with either `src_dir` or `out_dir`.

System variables, these are automatically resolved by Packer and can be used in other variables.

    :distro      // The name of the distribution
    :release     // The full release number being used for the template
    :arch        // The architecture of the OS being used for the template
    :image       // The name of the image being used for the template.
    :date        // The current datetime.
    :build_name  // The name of the build.

Standard variables:
    :name         // The name for Packer
    :out_dir      // The directory that the template output should be written to.
    :src_dir      // The source directory for files that should be copied with the template.
    :commands_src_dir  // The directory that command files are in.
    :http_dir      // The directory location for the "http" setting.
    :http_src_dir  // The source directory for files that will be copied to the "http_dir"
    :scripts_dir   // The directory location for the "scripts" setting
    :scripts_src_dir // The source directory for the files that will be copied to the "scripts_dir"

## Running Rancher
Build a Packer template for CentOS using the distro defaults:

	rancher build -distro=centos

Build a Packer template for Ubuntu using the distro defaults with an image override:

	rancher build -distro=ubuntu -image=desktop

Build a Packer template from a named build:

	rancher build 1204-amd64-server

Build a Packer template for Ubuntu using the distro defaults and from more than one named build:
	
	rancher build -distro=ubuntu -arch=i386 1404-dev 1204-amd64-server

## `rancher.toml` and Environment Variables
The `rancher.toml` file is the default core configuration file for Rancher. It contains the default locations for all of the TOML files that Rancher uses. Environment variables are supported. Rancher will first check to see if the environment variable is set. If it is empty, the relevant `rancher.toml` setting will be used.

For a full list of environment variables, please check the code comments.

The `RANCHER_CONFIG` environment variable can be used to specify a different file and location for Rancher's configuration, otherwise the `rancher.toml` file will be used by default.

## Rancher Configuration files
Information about defaults, supported distros, builds, and build lists are all stored in rancher configuration files, which are written in TOML. The defaults are set so that one can create a basic Packer template without any additional changes or overrides. 

For custom builds, only the differences between the default configuration and your desired configuration need to be set. Any missing values will use the default configuration for that build's distribution. These custom build configurations are stored in the rancher/conf.d/ directory.

If the defaults shipped with rancher are not to your liking, the default settings files, defaults.toml and supported.toml, can be modified. These files are found in the rancher/conf/ directory. Care should be taken when modifiying these settings.

The configuration files are described in order of precedence with later declarations overriding the prior ones.

### Settings
Rancher configuration settings take the form of `key=value` and Rancher parses that to a `key` and `value`, with the value taking the appropriate data type. This is space insensitive and it is assumed that the first `=` encountered split the `key` from the `value`. This means that any spaces within the value or any `=` are preserved, with the exception of leading and trailing spaces. Both the `key` and the `value` have their leading and trailing spaces trimmed.

For `boolean` values, `strconv.ParseBool()` is used with any error in parsing resulting in a value of `false`. Any value that `strconv.ParseBool()` evaluates to `true` are allowed to be used as values for true.

For `int` values, `strconv.Atoi()` is used to convert the value to an `int`. If the specified value results in an error from `strconv.Atoi()`, the error will be logged and processing will of that builder will stop.

#### `conf/defaults.toml`
The `defaults.toml` file contain most of the defaults for builds. A few settings that only apply at the distro level, like `base_url`, are not set here. The values in this file will be used, unless they are overridden later.

#### `conf/supported.toml`
The `supported.toml` file defines the supported operating systems, referred to as distros or distributions, and their settings, including any distro specific overrides. Because Rancher does distro specific processing, like ISO information look-up, Packer templates can be generated for supported distributions only. 

Each supported distro has a `defaults` section, which defines the default `architecture`, `image`, and `release` for the distro. These defaults define the target ISO to use unless the defaults are overridden by either flags or configuration.

### Distro Defaults
The settings in the `defaults.toml` and `supported.toml` files are merged and the result, for each supported distro, being the distribution's default template. This is used for Rancher builds that use thee `-distro` flag. They are also used as the basis for the templates defined in `builds.toml`. 

#### CentOS specific
For CentOS, the `base_url` should only be used if you want to use a specific iso redirect mirror. The mirror url should be the url download string, less the iso name. So an ISO located at `http://bay.uchicago.edu/centos/6.5/isos/x86_64/CentOS-6.5-c86_64-minimal.iso` would have a `base_url` of `http://bay.uchicago.edu/centos/6.5/isos/x86_64/`. No validation is done on the `base_url` so you need to make sure that the URL is correct for your release and architecture.

If the `base_url` is left empty, the url will be randomly selected from that isoredirect.centos.org page for the desired CentOS version and architecture.

#### `conf.d/builds.toml`
`builds.toml` contains all custom builds for Rancher. Each build has a name, which also becomes the name of the Packer template json file and the output directory for the resulting Packer template and its resources. Build specific settings and overrides are set here.

Each build, at minimum, must have a `distro`, which is the name of the distro for which the Packer template is being build. This is analogous to Packer's `type`.

The `settings` for a build are merged with the Rancher defaults and distribution's defaults. Any `_type` values replace the values in the prior definitions. Any `array` settings that are defined replace any prior definitions, no merging is done.

Builds are used in conjunction with the `build` command or are added to build lists in `buildlists.toml`.

#### `conf.d/buildlists.toml`
`buildlists.toml` defines named lists of builds. These name lists are used in conjunction with the `rancher run buildlistNames...` command. If you often find that the `build` sub-command is used with mupltiple build names, a buildlist might be useful instead. 

## Rancher Commands
Rancher uses mitchellh's cli package, with support for the `build` and `run` custom commands, along with the standard `version` and `help` commands.

### `build`
`rancher build <flags> buildNames...`

Supported Flags:

* -distro=<distro name>
* -arch=<architecture>
* -image=<image>
* -release=<release>

If the `-distro` flag is passed, a build based on the default setting for the distro will be created. The additional flags allow for runtime overrides of the distro defaults for the target ISO. This flag can be used in conjunction with named builds. If both the -distro flag is passed along with a space separated list of one or more named builds are passed to the `build` sub-command, both the default Packer template for the distro and all of the Packer templates for the passed build names will be created.

### `run`
`rancher run buildlistNames...`

Generate the Packer templates for all of the builds listed within the passed buildlists; buildlists is variadic. No flags are supported.

## Notes:
Each Packer section also has a `_type`, e.g. builders have a `builder_type`. This is a list of types that apply to the template being built. Each type must have a corresponding section defined. Only sections with a matching entry in the `_type` section will be processed by Rancher. These sections exist because the merged template may have more types defined than you want processed for a particular Packer template; e.g. there may be both a `vmware-iso` and a `virtualbox-iso` section defined in the final merged build template because those sections were part of the supported definition, but only the `vmware-iso` section is to be processed for this particular build.

For builders that require the `iso` information, you can specify your own information by populating the `iso_url` or `iso_urls`, `iso_checksum`, and `iso_checksum_type` settings. If these settings are not set, Rancher will look-up the information for you. For CentOS, this will result in a random mirror being chosen, unless you have specified the mirror in the `base_url` field.

Any command setting can either be set from a file or in the settings. If a file is going to be used, the filename must end in `.command` and be a valid file location. For `shutdown_command`, if it does not end in `.command`, Rancher assumes that the contents are the command. If `boot_command` exists in the `settings` section, it is assumed to represent a file, which must end in `.command`, and the `boot_command` array is populated from that file's contents. If the `boot_command` setting is not found in the settings section, the `arrays` section is checked and the array of boot commands found in the `boot_command` array are used. The `boot_command` setting in the settings section takes precedence over the array definition.

### File copy support
The only provisioner for which file copy is guaranteed to work properly is the `shell-scripts` provisioner as this is what Rancher was originally written to support. This functionality may work for other provisioners but is not guaranteed. Please file an issue for any other provisioners for which this should be supported. Or better yet, file a pull request!


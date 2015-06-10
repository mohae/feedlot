rancher
=======
>I am rarely happier than when spending entire day programming my computer to perform automatically a task that it would otherwise take me a good ten seconds to do by hand
> 
>   -Douglas Adams, _Last chance to See_

## About Rancher
Ranchers supply products to packers to process and package.  Rancher creates Packer templates for Packer to process and generate artifacts.  Ok, a bit of a stretch, but it's the best I could come up with and better than the other names I had thought of.

Rancher was created to make it easier to generate reproducable Packer templates and create updated Packer templates when new distro isos come out.

When creating Packer templates, Rancher will ensure that all of the resources required by that template are part of the generated Packer template.  The Rancher templates can contain information about the resources that they require and the actual resources will be located and made part of the Packer template.  If the referenced resource cannot be located, an error will occur and the requested Packer template will not be generated.

Rancher builds Packer templates from either Rancher build templates or, for quick flag based builds, from the supported distro defaults and Rancher build defaults.

## Why rancher
I originally started Rancher because I didn't like looking up the iso and checksum information for my Packer templates.  I also wasn't happy about replicating setting values between builders as I was too likely to make a mistake.  Overall, I found the process, though easy enough, more cumbersome than I would like.  I naively thought generating Packer templates from pre-defined configurations would be easier.  During which, I realized I'd have to support more than just builders, so Rancher evolved into generating a more complete template. 

The goal was to make it easier to create various configurations by minimally defining the differences between the named build and the distro's defaults; making it easier to generate updated templates as the releases that the templates depended on were updated.

I'm not sure if I achieved my goals, or if I just added complexity to different parts that didn't need to be there.

I chose TOML as the configuration file type, but there's no reason why YAML or something else couldn't be supported. It just hasn't been implemented.

## Rancher is cross-platform
Because Rancher uses Go, it can be compiled for cross-platform use.  It has been tested on Linux and Windows.

## Using Rancher
Currently, no pre-compiled versions of Rancher are made available.  To use Rancher, on must `go get` the repo and build it with `go build`.

    $ go get github.com/mohae/rancher
    $ cd GOPATH/src/github.com/mohae/rancher
    $ go build
	
You can either run the application from there, or move it somewhere in your path.  If you move `rancher`, make sure that your `out_dir` and `src_dir` settings are findable by the Rancher executable.  If you are using a custom application configuration file, make sure it's either in the application directory or the `RANCHER_CFG_FILE` environment variable is set with its location. 

Once built, Packer templates can be generated using the `rancher build` command.  This command accepts 0 or more build names and generates a Packer template for each named Rancher build template.  If no build name is passed, the `-distro` flag must be passed, at minimum. The `-distro` flag specifies which distro to use for Rancher's default build and generates a Packer template for that distro using that distro's defaults along with the Rancher defaults found in `defaults.toml`.  The specified distro must be supported.toml

### Basic Rancher command examples:  
Build a Packer template for CentOS using the distro defaults:

	rancher build -distro=centos

Build a Packer template for Ubuntu using the distro defaults with an image override:

	rancher build -distro=ubuntu -image=server

Build a Packer template from a named build:

	rancher build 1204-amd64-server

Build a Packer template for Ubuntu using the distro defaults and from more than one named build:
	
	rancher build -distro=ubuntu -arch=i386 1404-dev 1204-amd64-server

## Configuration  
The configuration directories and files are listed as they exist in the repo and represent the application default locations.  These files can be placed anywhere as long as the Rancher configuration has been updated, whether via the application configuration file or by environment variables.

### application configuration  
The `rancher.toml` file is the default core configuration file for Rancher.  It contains the default locations for all of the TOML files that Rancher uses. Environment variables are supported.  Rancher will first check to see if the environment variable is set.  If it is empty, the relevant `rancher.toml` setting will be used.

The `RANCHER_CONFIG` environment variable can be used to specify a different file and location for Rancher's configuration, otherwise the `rancher.toml` file will be used by default.

__Configuration file settings__  
Check the `rancher.toml.example` file for the config setting names and their application default values.

### Environment Variables  
Rancher supports using environment variables for configuration settings.  The environment variable name will always be upper-case and prefixed with `RANCHER_`.  The rest of the environment variable name will be the name of the configuration setting for which it applies. 

When an environment variable is set, it will override both the application default and configuration file values.  Only command-line flags can override environment variables.

#### Environment variable names for Rancher config files:

    RANCHER_CFG_FILE         // location of the rancher cfg file, e.g. rancher.toml
    RANCHER_SUPPORTED_FILE   // location of the file containing information on the supported distros
    RANCHER_DEFAULT_FILE     // location of the file containing Rancher's default build settings
    RANCHER_BUILD_FILE       // location of the file defining Rancher build templates
    RANCHER_BUILD_LIST_FILE  // location of the file defining sets of Rancher build templates

### `conf/`  
The `conf/` subdirectory holds the default configuration for Rancher builds, `default.toml`, and information about supported distros, `supported.toml`. 

#### `supported.toml`  
The `supported.toml` file defines the supported operating systems, referred to as distros or distributions, and their settings, including any distro specific overrides.  Because Rancher does distro specific processing, like ISO information look-up, Packer templates can be generated for supported distributions only and only for releases that that distribution currently supports.

Each supported distro has a `defaults` section, which defines the default `arch`, `image`, and `release` for the distro.  These defaults define the target ISO to use unless the defaults are overridden by either flags or configuration.

__Supported Distros:__   

    * CentOS
    * Debian
    * Ubuntu

#### `default.toml`  
The `default.toml` holds all the default settings for Rancher build templates.  A few settings that only apply at the distro level, like `base_url`, are not set here.  Unless a Rancher build template explicitly defines a setting, the values within the `default.toml` will be applied to all templates.  Ideally, Rancher build templates should only specify overrides to default settings and new settings, esp. Packer sections. 

The defaults are also used when Rancher templates are generated with a build template being specified:

TODO complete this example

### `conf.d/`  
The `conf.d/` subdirectory holds the configuration files for Rancher build templates and lists of builds.  These files are the ones that will be most commonly modified.

#### `build.toml`  
The `build.toml` contains named Rancher build templates.  A build template is a named specification for a Packer template and contains the settings and Packer sections that will apply to it.  When a Packer template is succesfully generated, the resulting `json` file, along with all resources, other than isos and certain files, will be copied to the output directory.  If a directory already exists in the target location and Rancher is set to archive prior builds, a compressed tarball will be created out of it, otherwise, the existing directory will be removed.  In either case, a new directory will be created and the artifacts of the Packer template will be placed within.

## Rancher build templates  
Rancher build templates, along with the underlying default and supported distro defaults, define what the resulting Packer template will consist of.  Each build template `builder`, `provisioner`, and `post-processor` section correspond to the Packer components in the same category.  In addition to these, Rancher templates also have some template settings and will have component type sections. 

### Rancher build template settings  
Rancher build template settings provide information that Rancher uses to help it create the build template's Packer template.  These settings usually only exist when there is a need to override the default setting.  Setting Rancher build template settings at the per build level also makes certain things more explicit.

`description`: corresponds to the Packer template _description_ setting.  
`name`: corresponds to the Packer template _name_ setting, by default, the build template name is used.  
`out_dir`: the directory to which the Packer template is written to and its resources copied to.  
`src_dir`: the directory which contains the source and resource files the build template references.  
`include_component_string`: a boolean as a string. Any value that Go's `strconv.ParseBool()` supports is allowed.   Any unsupported character results in this setting being evaluated to false. Please check the _notes_ section for more info.  
`min_packer_version`: corresponds to the Packer template *min_packer_version* setting.  

### Packer component type sections  
Each Packer section also has a `_type`, e.g. builders have a `builder_type`.  This is a list of types that apply to the template being built.  Each type must have a corresponding section defined.  Only sections with a matching entry in the `_type` section will be processed by Rancher.  These sections exist because the merged template may have more types defined than you want processed for a particular Packer template; by specifying the Packer section types that the build template will use the other definitions will be ignored.

The `provisioner_type` and `post-processor_type` sections are optional as their respective sections are optional in Packer.

### Rancher variable replacement  
Rancher supports a limited number of variables in the toml configuration files.  These are mostly used to allow for the path of files and names of things to be built based on other information within the configuration, e.g :distro is replaced by the name of the distro for which the Packer template is being built.  

Since Packer uses Go's template engine, Rancher variables are prefixed with a colon, `:`, to avoid collision with Go template conventions, `{{}}`.  Using a `:` in non-variable values may lead to unexpected results, as such don't use it unless you are prefixing a variable.  A different delimiter can be specified in the Rancher configuration file by modifying the `param_delim_start` setting.  By using a different delimiter for Rancher, your build templates can still contain Packer variables.

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

For resource locations, Rancher will attempt to locate the specified file by searching various possible locations. This is covered in the _Find Algorithm for Build Template Sources and Resources_ section. For resources that will be part of the Packer template, their name should reflect what the Packer template will be using and not the path where Rancher can find it.

### Build template Settings
Rancher build configuration settings take the form of `key=value` and Rancher parses that to a `key` and `value`, with the value taking the appropriate data type. This is space insensitive and it is assumed that the first `=` encountered split the `key` from the `value`. This means that any spaces or equals, `=`, within a value are preserved, with the exception of leading and trailing spaces. Both the `key` and the `value` have their leading and trailing spaces trimmed.

For `boolean` values, `strconv.ParseBool()` is used with any error in parsing resulting in a value of `false`. Any value that `strconv.ParseBool()` evaluates to `true` are allowed to be used as values for true.

For `int` values, `strconv.Atoi()` is used to convert the value to an `int`. If the specified value results in an error from `strconv.Atoi()`, the error will be logged and processing will of that builder will stop.

#### Command files
Command settings, like `boot_command` and `shutdown_command`, support the use of command files by specifying the command file in the setting value, instead of the actual command string. Any command setting value that ends in `.command` will be assumed to reference a Rancher command file. The setting will be populated from the referenced file. If the setting only supports a single line, the first line of the command file will be used. For settings that support arrays of commands, like `boot_command`, the entire contents of the file will be used as the commands.

Rancher's algorithm for finding build template sources applies to command files.

#### Find Algorithm for Build Template Sources and Resources
Packer templates can contain resources that it references, e.g. `http_dir` and scripts for the shell provisioner. The location specified in the templates are not where Rancher would find the resource files to copy into the resulting Packer template. 

Rancher templates may contain references to sources from which it obtains Packer template data. These sources are mainly command files that contain shell commands to be inserted into the generated Packer template.

In order to be used, both referenced resources and sources must be findable by Rancher. This algorithm should make it possible for files to be referenced in multiple Packer templates along with an easy way of restricting the scope, or availablility of files to certain templates or other category.

The `src_dir:` setting is used to specify the root location for Rancher resources; all searches for a resource location will be restricted to this path. The base `src_dir` is defined in `default.toml`; it can be overridden by a Rancher build template. The  resource path, relative to the `src_dir` is the last location that Rancher will check for the specified resource before returning a `not found error`.

Except for searches that include the `build_name`, Rancher tends to search from deeper in the directory try to higher up, making it easier to scope resource availability, when useful. Build name based searches start with the root and work their way down because `build_name`s must be unique, and as such, are already specific.

The first match encountered is returned. To provide help in understanding how Rancher locates resources, the paths searched can be logged out by setting the `log_level_file` or `log_level_stdout` to `TRACE. `

The base search order is:

    :src_dir/build_name
	:src_dir/distro/build_name
    :src_dir/distro/release/build_name
    :src_dir/distro/releaseBase/build_name
    :src_dir/distro/release/arch
    :src_dir/distro/releaseBase/arch
    :src_dir/distro/release/
    :src_dir/distro/releaseBase/
    :src_dir/distro/arch
    :src_dir/distro/
    :src_dir/

`:src_dir` is the value for the `src_dir` setting in the Rancher build template.
`arch` is the iso architecture for the Packer template.
`build_name` is the name of the Rancher build template that is being used to generate the Packer template.
`distro` is the name of the distro for which this template is being built, e.g. `ubuntu`, `centos`.
`release` is the full release number for the build, e.g.  `6,6`.
`releaseBase` is the whole version number for the build, e.g. the `releaseBase` of `6.6` is `6`.

##### Component name handling
When applicable, resource searches are first performed by prepending the requested resource path with the name of the Packer component for which this resource applies prior to searching for the file, e.g. a search for `vagrant.sh` for the Shell provisioner would use `shell/vagrant.sh'`. If a match isn't found, another search will be done without the component name prefixed, i.e. `vagrant.sh`.

If the Packer component name is hyphenated, an intermediate search will be done between the component name search and the base search. This intermediate search will use the first part of the hyphenated name as the parent directory to the requested resource, e.g. `amazon` in addition to `amazon-ebs`.

Supporting component names in the search path allow for the resources to be specified within a build template in a manner that's more useful for Packer templates while still supporting segregation of source files by usage, e.g. shell scripts in the shell directory, ansible playbooks in ansible, etc. helping to keep them organized on the source side.'

#### Example search start-points using the `chef-solo` Packer component.
If searching for the `chef.cfg` file, Rancher would do up to three searches using the above search paths.

The first search would use the full component name as the parent directory of `chef.cfg`. It's first search would be:  `src_dir/build_name/chef-solo/chef.cfg`.

If `chef-solo/chef.cfg` was not found in any of the searched directories, Rancher would then look for the file using the Packer component's base name, `chef`, as the parent directory. This only applies to Packer components that have a hyphenated name. The second search would start at `src_dir/build_name/chef/chef.cfg`

If `chef/chef.cfg` was not found ini any of the searched directories, Rancher would finally look for the resource with just it's path. The final search would start at `src_dir/build_name/chef.cfg`

## Packer templates
The output of a Rancher build is 1 or more Packer templates and their resources,

### Supported Packer Components
Each supported Packer component has various configuration options. Any required configuration option is supported and a missing required element will result in a processing error, with the information logged to the Log. Most optional configuration options are supported, any that are not will be listed under as a `not supported configuration option` in that section's comments. If there are any options defined in the template that either do not exist for or are not supported by that Packer section type, a warning message will be logged, but the processing of the template will continue.

Some supported Packer section types may have unsupported settings. Usually these are settings that contain an object, though some others may not be supported either. For detailed information, please check the docs.

#### Supported Distro
    * centos - support for setting ISO information for 7 not implemented.
    * debian
    * ubuntu

#### Supported Builders
    * amazon-ebs
    * digitalocean
    * docker
    * googlecompute
    * null
    * virtualbox-iso
    * virtualbox-ovf
    * vmware-iso
    * vmware-vmx
    
Rancher also has a `common` builder which can contain settings which may exist in more than 1 builder, e.g. `iso_checksum`. Rancher will merge all `common` builder settings with each defined builder. In situations where a key exists in both the `common` builder and a supported builder, the supported builder's value will apply. Any settings that are defined in the `common` builder will be ignored if a supported builder does not support that key.

#### Supported Post-processors
    * compress
    * docker-import
    * docker-push
    * docker-save
    * docker-tag
    * vagrant
    * vagrant-cloud
    * vsphere

#### Supported Provisioners 
    * ansible-local
    * chef-client
    * chef-solo
    * file-uploads
    * salt-masterless
    * shell


## Rancher Commands
###Rancher uses mitchellh's cli package, with support for the `build` and `run` custom commands, along with the standard `version` and `help` commands.

### `build`
`rancher build <flags> buildNames...`

Supported Flags:

* -distro=<distro name>
* -arch=<architecture>
* -image=<image>
* -release=<release>

If the `-distro` flag is passed, a build based on the default setting for the distro will be created. The additional flags allow for runtime overrides of the distro defaults for the target ISO. This flag can be used in conjunction with named builds. If both the -distro flag is passed along with a space separated list of one or more named builds are passed to the `build` sub-command, both the default Packer template for the distro and all of the Packer templates for the passed build names will be created.

## Notes:
### `include_component_string`
    default: true

The `include_component_string` is a Rancher build setting that controls whether or not the name of the Packer component for which the resource will be used should be prepended as the parent directory to that resource in the Packer template. 

This is useful for keeping various packer resources in separate directories by usage. It also makes specifying the resource within the build template easier. One drawback to this is that the component name may not be the desired value, but it will still work, e.g. `scripts` is a more common name for the directory within a Packer template containing the shell scripts, using the component name, the directory would be `shell` instead.

__Example:__
In a Rancher build template, a list of shell scripts may be:
    [provisioners.shell.arrays]
		scripts = [ 
			"setup.sh",
			"vagrant.sh",
			"sudoers.sh",
			"cleanup.sh",
		]
    
With `include_component_string` set to true, the Packer template section would look like:
    "provisioners": [
     	{
      		"type": "shell",
			"scripts": [
				"shell/setup.sh",
				"shell/vagrant.sh",
				"shell/sudoers.sh",
				"shell/cleanup.sh"
			]
		}
	]

And the files would be copied to:
    packer_template/shell/setup.sh
	packer_template/shell/vagrant.sh
	packer_template/shell/sudoers.sh
	packer_template/shell/cleanup
	
With `include_component_string` set to false, the Packer template section would look like:
    "provisioners": [
     	{
      		"type": "shell",
			"scripts": [
				"setup.sh",
				"vagrant.sh",
				"sudoers.sh",
				"cleanup.sh"
			]
		}
	]

And the files would be copied to:
    packer_template/setup.sh
	packer_template/vagrant.sh
	packer_template/sudoers.sh
	packer_template/cleanup

### Specifying your own iso information
For builders that require the `iso` information, you can specify your own information by populating the `iso_url` or `iso_urls`, `iso_checksum`, and `iso_checksum_type` settings. If these settings are not set, Rancher will look-up the information for you. For CentOS, this will result in a random mirror being chosen, unless you have specified the mirror in the `base_url` field.

### Key handling
While Rancher ensures that most referenced resources are copied to the resulting template, this does not apply to key files. When generating Packer templates, Rancher does not attempt to locate any referenced key files. It is assumed that these keys will be in the location specified in the template. 

Please file an issue for any key resources that may be copied during Packer template generation from Rancher builds.toml

## Issues: questions, bugs, changes, etc.
Please file an issue for an issue for any questions, bugs, changes, wishes, etc. that you may have related to Rancher. Better yet, submit a pull request when applicable!

# feedlot-example
Example configurations, templates, and resources for feedlot; can be used when running [feedlot](https://github.com/mohae/feedlot) with the `-example` flag; or with feedlot configured with `example=true`.

## About
This repository contains examples for feedlot that can be used to either generate example Packer templates or as a basis for your own customized files..

### Generating Example Packer Templates
When creating your own Packer templates, it can be helpful to see examples. Feedlot can be used for this, even if you don't use feedlot to generate Packer Templates as part of your toolchain.

Currently, the `default` file contains only the required settings for each Packer component.

The `required` configuration file has builds defined for each supported distro with all Packer components specified. The resulting Packer template will contain all Packer components and only their required setting.

The `build` configuration file has builds defined for each supported distro using the `virtualbox-iso` and `vmware-iso` builders, the `shell-script` provisioner, and the `vagrant` post-processor. This is a common combination for Packer users that wish to generate artifacts for local development.

_Additional examples will be added._

To generate example Packer templates, use feedlot's `-example` flag, or set the `feedlot` config file's `example` setting to `true`.

The other role of these examples is to provide examples of how feedlot builds can be defined so that you can define your own feedlot build templates. While not all supported scenarios are not included, it does provide a starting point to simplify, hopefully, the process of creating customized builds.

The examples also contain example source files that can be used. Currently, only a few example sources exist.

When run in example mode, feedlot does not require files to exist. If it cannot locate a specified file, it concatenates the Packer component (builder, post-processor, or provisioner) and the specified file path and uses that value in the template.

## feedlot conf files and enviornment vars
The `feedlot.json` and `feedlot.toml` files contain all of the feedlot supported configuration settings with feedlot's default values, with the exception of the `conf_dir` and `format` settings.

Each file's `format` setting is set to match the format of the configuration file; i.e. `feedlot.json`'s format is `json` and `feedlot.toml`s format is set to `toml`. In addition, each file's `conf_dir` is set to match the format specified within that file.  It is possible to mix these, e.g. `feedlot.json`'s format can be set to `toml`, in which case feedlot will look for the `toml` version of the files.

The `conf/toml` and `conf/json` subdirectories exist to keep the different versions separate. Normally, one would have those files in `conf` and the `feedlot` config's `conf_dir` setting would be set to `conf`, which is feedlot's application default.

The configuration file is optional. If it is missing, an error will not occur. Instead, feedlot will use the application's defaults along with any environment variables that are set and any flags that are passed: in that order of precedence.

Alternatively, if you prefer environment variables, you can use those. Environment variables take the form of `FEEDLOT_` + setting name; e.g. `FEEDLOT_CONF_DIR` for the `conf_dir` setting. These variables follow the uppercase convention. If an environment variable is set, it will take precedence over any value found in the feedlot conf file and can only be overridden by a flag.

Finally, any feedlot conf file or  can be overridden by passing a flag at runtime.

## Build configuration files
feedlot uses several configuration files for defining information about supported distros, defaults for both feedlot and various Packer components, custom build templates, and lists of builds that can be run.

### `supported`
The `supported` configuration file defines what distros are supported along with details about what is supported for that distro, e.g. releases or versions, architectures, iso images, etc. The default release, architecture, and iso image for each supported distro is also defined here.

If a distro is not in the `supported` configuration file, support for it needs to be added to feedlot.

Adding support for various versions, or releases, of a distro may or may not require modifications to feedlot. For example, CentOS changed the way they structure their versions and name their releases with CentOS 7, which required changes to feedlot. On the other hand, adding a new release to Ubuntu only requires adding the release to the `release` array.

If there is a new release and it hasn't been added to the `supported` configuration, please file an issue, or make a pull request.

#### supported distros
Feedlot supports the current releases for versions that each distro supports.

* CentOS
* Debian
* Ubuntu

### `default`
The `default` configuration file defines defaults for builds. Any setting that will be consistent across builds should be defined here so that each build doesn't need to define it.

Anything set in this file can be overridden by a feedlot build template.

### `build_list`
The `build_list` configuration file defines names lists of build templates. These are used with the `feedlot run` command to generate multiple Packer templates without having to explicitly list all of the templates you wish to build.

_This is not currently implemented._

### `build`
The `build` configuration file defines various feedlot build templates that can be used to generate Packer templates. These builds define any overrides or additions to the defaults, along with what Packer components they will build.

In adddition to the `build` configuration file, any files found in the conf directory that aren't the `build_list`, `default`, or `supported` configuration files, and that are of the same type as the supported format, will be read as additional `build` configuration files. This enables you to separate out builds to different files. The `required` configuration file is an example of this.

__Note:__ build names should be unique. While this isn't enforced, in the case of more than one build having the same name, feedlot will use the first match found. There is no guarantee as to which one this ends up being.

## `packer_sources`
The `packer_sources` directory contains example source files for builds. During a Packer template build, feedlot will find the sources for the resources referenced in the template. If feedlot is in example mode, errors will not be generated for missing sources; regular feedlot builds will error on missing resources.

These example sources are valid and can be used or customized to fit your needs. These sources are not exhaustive and additional source files may be added in the future. Feel free to contribute additional sources (though this is not meant to be an exhaustive resource as these are examples.)

### `commands`
Feedlot can extract commands from command files, files ending in `.command`, for settings that use commands, e.g. `boot_command` and `execute_command`. Alternatively, these commands can be specified inline like one would do for Packer templates.

### `http`
The http directory contains files that will be copied by Builders that use the `http_directory` setting, e.g. `virtualbox-iso` and `vmware-iso`.  The `.cfg` files are valid for their respective distros. You may want to modify the files to meet your specific needs.

### Missing source files
An error will not occur if feedlot is in example mode and the source file is not found. If feedlot is not in example mode, a missing source file will result in an error.

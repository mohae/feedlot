# rancher-example
Example configurations, templates, and resources for Rancher; can be used when running [Rancher](https://github.com/mohae/rancher) with the `-example` flag; or with Rancher configured with `example=true`.

## About
This repository contains examples for Rancher that can be used to either generate example Packer templates or as a basis for your own customized files..

### Generating Example Packer Templates
When creating your own Packer templates, it can be helpful to see examples. Rancher can be used for this, even if you don't use Rancher to generate Packer Templates as part of your toolchain.

Currently, the `default` file contains only the required settings for each Packer component. 

The `required` configuration file has builds defined for each supported distro with all Packer components specified. The resulting Packer template will contain all Packer components and only their required setting.

The `build` configuration file has builds defined for each supported distro using the `virtualbox-iso` and `vmware-iso` builders, the `shell-script` provisioner, and the `vagrant` post-processor. This is a common combination for Packer users that wish to generate artifacts for local development.

_Additional examples will be added._

To generate example Packer templates, use Rancher's `-example` flag, or set the `rancher` config file's `example` setting to `true`.

The other role of these examples is to provide examples of how Rancher builds can be defined so that you can define your own Rancher build templates. While not all supported scenarios are not included, it does provide a starting point to simplify, hopefully, the process of creating customized builds.

The examples also contain example source files that can be used. Currently, only a few example sources exist.

## Rancher config files
The `rancher.json` and `rancher.toml` files contain all of the Rancher supported configuration settings with Rancher's default values, with the exception of the `conf_dir` and `format` settings. 

Each file's `format` setting is set to match the format of the configuration file; i.e. `rancher.json`'s format is `json` and `rancher.toml`s format is set to `toml`. In addition, each file's `conf_dir` is set to match the format specified within that file.  It is possible to mix these, e.g. `rancher.json`'s format can be set to `toml`, in which case Rancher will look for the `toml` version of the files.

The `conf/toml` and `conf/json` subdirectories exist to keep the different versions separate. Normally, one would have those files in `conf` and the `rancher` config's `conf_dir` setting would be set to `conf`, which is Rancher's application default.

This config file is optional. If it is missing, an error will not occur. Instead, Rancher will use the application's defaults. 

Alternatively, if you prefer environment variables, you can use those. Environment variables take the form of `rancher_` + setting name; e.g. "rancher_conf_dir" for the `conf_dir` setting. If an environment variable is set, it will take precedence over any value found in the Rancher config file.

Lastly, any setting in the Rancher config file can be overridden by passing a flag at runtime.

## Build config files
Rancher uses several config files for defining information about supported distros, defaults for both Rancher and various Packer components, custom build templates, and lists of builds that can be run.

### `supported`
The `supported` config file defines what distro's are supported along with details about what is supported for that distro, e.g. releases or versions, architectures, iso images, etc. The defaults release, architecture, and iso image for each supported distro is also defined here.

If a distro is not in the `supported` config file, support for it needs to be added to Rancher.

Adding support for various versions, or releases, of a distro may or may not require modifications to Rancher. For example, CentOS changed the way they structure their versions and name their releases with CentOS 7, which required changes to Rancher. On the other hand, adding a new release to Ubuntu only requires adding the release to the `release` array.

If there is a new release and it hasn't been added to the `supported` configuration, please file an issue, or make a pull request.

### `default`
The `default` config file defines defeaults for builds. Any setting that will be consistent across builds should be defined here so that each build doesn't need to define it.

Anything set in this file can be overridden by a Rancher build template.

### `build_list`
The `build_list` config file defines names lists of build templates. These are used with the `rancher run` command to generate multiple Packer templates without having to explicitely list all of the templates you wish to build.

_This is not currently implemented._

### `build`
The `build` config file defines various Rancher build templates that can be used to generate Packer templates. These builds define any overrides or additions to the defaults, along with what Packer components they will build.

In adddition to the `build` config file, any files found in the conf directory that aren't the `build_list`, `default`, or `supported` config files, and that are of the same type as the supported format, will be read as additional `build` config files. This enables you to separate out builds to different files. The `required` config file is an example of this.

__Note:__ build names should be unique. While this isn't enforced, in the case of more than one build having the same name, Rancher will use the first match found.

## `packer_sources`
The `packer_sources` directory contains example source files for builds. During a Packer template build, Rancher will find the sources for the resources referenced in the template. If Rancher is in example mode, errors will not be generated for missing sources; regular Rancher builds will error on missing resources.

These example sources are valid and can be used or customized to fit your needs. These sources are not exhaustive and additional source files may be added in the future. Feel free to contribute additional sources (though this is not meant to be an exhaustive resource as these are examples.)

### `commands`
Rancher can extract commands from command files, files ending in `.command`, for settings that use commands, e.g. `boot_command` and `execute_command`. Alternatively, these commands can be specified inline like one would do for regular Packer templates.

### `http`
The http directory contains files that will be copied by Builders that use the `http_directory` setting, e.g. `virtualbox-iso` and `vmware-iso`.  The `.cfg` files are valid for their respective distros. You may want to modify the files to meet your specific needs. 

### Missing source files
An error will not occur if Rancher is in example mode and the source file is not found. If Rancher is not in example mode, a missing source file will result in an error.

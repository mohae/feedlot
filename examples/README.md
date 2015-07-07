# rancher-example-tpls
Example templates for Rancher; can be used with `rancher example ...`

## About
This repository contains examples for both [Rancher](https://github.com/mohae) build templates and [Packer](https://packer.io) templates.

#### Packer components
A component of Packer processes a Packer Builder, Post-processor, or Provisioner.  During a Rancher build, only the Packer components specified by the {{component_type}} field are processed.

##### Supported Component Types
__Builders__:  `builder_type`  
    * `virtualbox-iso`
    * `virtualbox-ovf`
    * `vmware-iso`
    * `vmware-vmx`

__Post-processors__: `post_processor_type`  
    * vagrant

__Provisioners__: `provisioner_type`  
    * shell

### Packer examples
The example Packer templates are contained within the `packer/` directory. These examples are intended to make it easier for people new to Packer to understand Packer templates. These templates are not guaranteed to work; some of their files may not be up to date or may not be consistent with the image's operating system; e.g. uses preseed.cfg instead of cfg.ks.

All of the Packer examples are generated using Rancher and the example configs.

### Rancher examples
The example repository also contains example configuration files in both `json` and `toml` formats. The configuration between the formats are not guaranteed to be consisistent. 

#### JSON
The example builds in `conf.d/example.build.json` either contain only the required elements for that Packer component or what was documented in the component's online documentation. In the case where there are multiple required settings that are mutually exclusive, only the first example was used.

#### TOML

### Src files
The source files are mostly empty. The command files contain the commands from the Packer component 

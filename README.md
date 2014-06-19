rancher
=======

Generate Packer Templates

Rancher supports VirtualBox and VMWare builders for Packer. This includes allowing for common builder settings and builder specific overrides.
Only the vagrant post-processor is currently supported. Not all vagrant settings are supported, just the minimal to get things working, for now. Overrides are not supported.
Only the shell provisioner is currently supported. Overrides are not supported.

Supported Distros:
    * ubuntu

Configuration files: All rancher configuration files are written in TOML. The defaults are set so that one can create a basic Packer template without any additional changes or overrides. For custom builds, only the differences between the default configuration and your desired configuration need to be set. Any missing values will use the default configuration for that build's distribution. These custom build configurations are stored in the rancher/conf.d/ directory.
If the defaults shipped with rancher are not to your liking, the default settings files, defaults.toml and supported.toml, can be modified. These files are found in the rancher/conf/ directory. Care should be taken when modifiying these settings.

Issues:

 * The Run command is not supported, though stubbed out.
 * Some comments are outdated and need to be revised accordingly
 * Some tests are buggy resulting in manual cleanup of the rancher/test_files/out directory.
 * No pressed.cfg for ubuntu
 * Command and shell scripts for ubuntu may not be totally correct.

# Building the Pure1 Unplugged Appliance
The base for this is CentOS. We use a few tools to help.

# TL;DR:

Run
```bash
./scripts/build/build_installer_iso_in_docker.sh
```

This goes off and creates a special CentOS docker image (as needed) and then inside of it uses Lorax to construct a fresh
minimal image with only our stuff in it. Output from this script is in `<repo root>/bin/iso`. The resulting ISO can be
booted from and used to install the Pure1 Unplugged appliance.

# Lorax
We build the image using [Lorax](https://weldr.io/lorax/lorax.html). There is a docker image and some helpers
in [../images/lorax-build](../images/lorax-build).

# lorax-builder image
Build it with the helper via

```bash
./scripts/build/build_lorax_image.sh
```

This image has a copy of all the tooling required.

_Note:_ To build appliance images (.iso, .qcow2, etc) you will need to run it as privileged! It doesn't seem to be
able to create loopback devices without it...

# CentOS Base
The version of CentOS we are building is based 100% off the docker image version, more specifically the version it
is using with its yum repositories. Lorax will pull from the configured repos to grab the OS and kernel. See the
`lorax-build` image for details. To update just change the version specified.

# Customizing CentOS
We need to pre-install some packages (docker, etc) and then drop our content into the image. This is done with a
combination of Lorax templates, RPM voodoo, kickstart files, and a little luck.

From a high level there are a few steps involved...

1) Download all the RPMs we will eventually need on our system into our tmp build dir (see [rpm-list.txt](./rpm-list.txt)).
2) Create a yum repo from the directory
3) Create an RPM containing the repo _and_ some config files we need, this ones called `pure1-unplugged-boot-config`
4) Create _another_ yum repo with only our RPM in it
5) Tell lorax via cli arg to use the repo from (4) as a "source" for the image
6) Lorax (while processing our template) will install our `pure1-unplugged-boot-config` RPM, which unpacks the big yum repo into the installtree
7) That same template will install our configs (kickstart files, etc)
8) Another lorax template (our `arch` template) then will copy the installed repository from the installtree to the special
`iso-graft` dir. This one gets grafted into the installer iso.

At boot time the anaconda image is loaded and run, this linux live image is what was built from the templates in step 6/7.
It defaults to using our kickstart file which in turn points to the "cdrom" image we smuggled RPMs into (same way the "all-in-one" DVD's do).

## Background Reading

* [https://weldr.io/lorax/](https://weldr.io/lorax/)
* [https://weldr.io/lorax/lorax.html#custom-templates](https://weldr.io/lorax/lorax.html#custom-templates)
* [https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/7/html/installation_guide/sect-kickstart-syntax](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/7/html/installation_guide/sect-kickstart-syntax)

# Upgrading RPM's
1. Install from the ISO as normal (including deployment and everything required to be up and running)
1. Add the external CentOS mirrors (break our offline-only rule), if needed snag the repo list from a normal CentOS version.
1. Add the kubernetes repo (see lorax image build for reference)
1. Ensure all the required packages are installed by doing something like `cat rpm-list.txt | xargs yum install -y`
   * If there are errors it likely means we need to upgrade some system packages, fix them and don't ignore them!
1. Do a normal YUM update for the package *WARNING: DO NOT UPDATE `docker-ce` UNLESS REQUIRED! IT IS PINNED TO AN OLDER VERSION, SEE [https://kubernetes.io/docs/setup/cri/#docker](https://kubernetes.io/docs/setup/cri/#docker)* 
1. Re-gen the `rpm-list.txt` file by running:
   ```bash
    rpm -qa | sort | awk '!a[$0]++' > rpm-list.txt
    ```
    This gets all installed RPMs, sorts them by name, and then removes any duplicates
1. Save the new rpm-list.txt in the repo
_Note_: If any new repositories are required, ensure that they have been added to the `lorax-build` Docker image so
that the RPM's can be downloaded.

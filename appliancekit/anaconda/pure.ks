# This kickstart file should be used instead of the default anaconda interactive one

# Installation files source, force the cdrom (from our ISO)
install
cdrom

firstboot --enable
eula --agreed

# Basic auth setup
auth --enableshadow --passalgo=sha512

# Everyone has a us keyboard?
#keyboard --vckeymap=us --xlayouts='us'

# Setting up language to English
lang en-US.UTF-8

# Default to something..
timezone America/Los_Angeles --nontp --utc

# Setting up MBR
zerombr
bootloader --location=mbr --boot-drive=sda

# Setting up Logical Volume Manager and autopartitioning
ignoredisk --only-use=sda
clearpart --all --drives=sda --initlabel
autopart --fstype=ext4 --nohome --type=lvm

# Configure the bootloader, override the MBR
bootloader --location=mbr --boot-drive=sda

# Setting up firewall and enabling SSH for remote management
firewall --enabled --service=ssh

# Setup some starter services
services --enabled=sshd

# Setting up Security-Enhanced Linux into permissive, as required by kubernetes
# https://kubernetes.io/docs/setup/independent/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl
selinux --permissive

# Eject cdrom and reboot when we are done installing
reboot --eject

# Installing only packages we need for pure1-unplugged, these need to all be available from cdrom
# The rpm-comps.xml defines the Pure1 Unplugged environment, specified below.
%packages
@^Pure1 Unplugged
%end

%addon com_redhat_kdump --enable --reserve-mb='auto'
%end

# Get rid of the standard yum repositories, ours get added in later
# but avoid any issues between OS install and pure1-unplugged setup
%post
!/bin/bash
rm /etc/yum.repos.d/*.repo
%end

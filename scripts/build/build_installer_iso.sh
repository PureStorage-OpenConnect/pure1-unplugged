#!/usr/bin/env bash

set -xe

# !! This needs to be run in CentOS, matching the targeted version !!
if [[ -z "$(grep CentOS /etc/*-release)" ]]; then
    echo "ERROR: This script MUST be run on CentOS to build the CentOS based ISO."
    exit 1
fi

# Expected to be passed in by caller (usually docker)
REPO_ROOT=${CONTAINER_REPO_ROOT}

if [[ -z "${REPO_ROOT}" ]]; then
    echo "Must set CONTAINER_REPO_ROOT env var!"
    exit 1
fi

VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

OS_NAME="Pure1-Unplugged"
OS_VERSION=${VERSION} # based off git versioning
OS_REVISION=${VERSION}  # based off git versioning

BUILD_DIR=${REPO_ROOT}/build/iso
BUILD_TMP_DIR=${BUILD_DIR}/tmp

if [[ -d "${BUILD_DIR}" ]]; then
    rm -rf "${BUILD_DIR}"
fi
mkdir -p ${BUILD_DIR}
mkdir -p ${BUILD_TMP_DIR}

# We need to customize the repo file for the build to match the version of CentOS
# the image is based on. If we use the default variables in it we get the target
# versions substitued which causes problems..
cp /etc/yum.repos.d/CentOS-Base.repo ${BUILD_TMP_DIR}/centos7-base.repo
sed -i s/.releasever/7/g ${BUILD_TMP_DIR}/centos7-base.repo

# We need to change the base URL to use the CentOS vault.
sed -i "s/#baseurl=http:\/\/mirror.centos.org\/centos\/7/baseurl=http:\/\/vault.centos.org\/${CENTOS_VERSION}/g" ${BUILD_TMP_DIR}/centos7-base.repo
# Disable the mirror list because the mirrors lie (they make it so we can't find packages for some reason)
sed -i "s/mirrorlist/#mirrorlist/g" ${BUILD_TMP_DIR}/centos7-base.repo

# Copy this back to use our updated version
cp ${BUILD_TMP_DIR}/centos7-base.repo /etc/yum.repos.d/CentOS-Base.repo

# Clean all just to remove any caches we may somehow have
yum clean all

# Generate our boot config RPM, which is in itself installing along with some extra configs
mkdir -p ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm
mkdir -p ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/Packages
cat ${REPO_ROOT}/appliancekit/rpm-list.txt | xargs yumdownloader -y --destdir=${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/Packages

# Copy in the pure1-unplugged RPM to our repo so we can package it in the image
cp ${REPO_ROOT}/build/rpm/pure1-unplugged-*.rpm ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/Packages

# Generate the repo from the "Packages" and our comps.xml (group metadata required by anaconda and our kickstarter)
createrepo \
    -v \
    -g ${REPO_ROOT}/appliancekit/rpm-comps.xml \
    -o ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm \
    ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/

# Last but not least sneak in our kickstart file and anaconda image resources
cp -r ${REPO_ROOT}/appliancekit/anaconda ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm

# Now the wrapper RPM
fpm \
    --workdir ${BUILD_TMP_DIR} \
    -s dir \
    -t rpm \
    -n pure1-unplugged-boot-config \
    -v ${VERSION} \
    -m "support@purestorage.com" \
    --url "pure1.purestorage.com" \
    -p ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-${VERSION}-x86_64.rpm \
    ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/Packages=/ \
    ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/repodata=/ \
    ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-rpm/anaconda=/opt/pure1-unplugged/

# now the kinda weird part, setup *another* repository with our boot config rpm in it. We then tell lorax to use it
# as a source (-s) so we can yum install our boot-config rpm into the boot media chroot'ed yum base.
LORAX_REPO=${BUILD_TMP_DIR}/lorax-repo
mkdir -p "${LORAX_REPO}"
cp ${BUILD_TMP_DIR}/pure1-unplugged-boot-config-*.rpm ${LORAX_REPO}
createrepo ${LORAX_REPO}

cat <<EOF > ${BUILD_TMP_DIR}/lorax.repo
[lorax]
name=Lorax Temporary Repo
baseurl=file://${LORAX_REPO}
gpgcheck=0
enabled=1
EOF

# This volume/filesystem label is important! It needs to match up with what is in the udev rules
# we install in the node (14-pure1-unplugged-media-by-label-auto-mount.rules) because it is how we identify
# the ISO and auto-mount it for using it as a yum repo. Both must be kept in sync.
PURE1_UNPLUGGED_VOLID="${OS_NAME}_x86_64"

# And finally, call lorax to generate our boot iso
lorax \
    -p "${OS_NAME}" \
    -v "${OS_VERSION}" \
    -r "${OS_REVISION}" \
    --repo "${BUILD_TMP_DIR}/centos7-base.repo" \
    --repo "${BUILD_TMP_DIR}/lorax.repo" \
    --isfinal \
    --add-template ${REPO_ROOT}/appliancekit/pure-lorax.tmpl \
    --add-arch-template ${REPO_ROOT}/appliancekit/pure-lorax-arch.tmpl \
    --logfile ${BUILD_DIR}/lorax.log \
    --tmp ${BUILD_TMP_DIR} \
    --bugurl "https://support.purestorage.com/" \
    --volid ${PURE1_UNPLUGGED_VOLID} \
    --buildarch "x86_64" \
    --rootfs-size 10 \
    ${BUILD_DIR}/lorax-results

cp ${BUILD_DIR}/lorax-results/images/boot.iso ${BUILD_DIR}/${OS_NAME}-${VERSION}.iso
pushd ${BUILD_DIR}/
sha1sum -b ${OS_NAME}-${VERSION}.iso > ${OS_NAME}-${VERSION}.iso.sha1
popd

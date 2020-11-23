ARG CENTOS_VERSION="7.6.1810"
FROM centos:${CENTOS_VERSION}
# ^^ the docker image version *must* match the version of centos we are building!!

# Hacky junk to make docker see the ARG after the FROM, the value is the same as above
ARG CENTOS_VERSION
ARG KUBERNETES_VERSION

# export this for use in scripts building images
ENV CENTOS_VERSION=${CENTOS_VERSION}

# Setup our repos required for kubernetes
COPY images/lorax-build/kubernetes.repo /etc/yum.repos.d/
RUN yum-config-manager \
        --add-repo \
        https://download.docker.com/linux/centos/docker-ce.repo

RUN yum update -y && yum install -y \
    lorax \
    rpmdevtools \
    ruby-devel \
    gcc \
    make \
    rpm-build \
    rubygems \
    createrepo \
    yum-utils \
    kubeadm-${KUBERNETES_VERSION} \
    docker-ce \
    git

# Note we install kubeadm only so we can snag a list of its required images with the config shipped
# for this version of kubeadm (matches the one we install with our installer iso).

RUN gem install --no-ri --no-rdoc fpm -f

COPY images/lorax-build/lorax.conf /etc/lorax/

# Apply a patch to lorax to increase the size of the disk (so we can fit in our RPMs)
COPY images/lorax-build/lorax.patch /root/
RUN patch --strip=0 --directory=/usr/lib/python2.7/site-packages/ </root/lorax.patch

# Add helm to the image so we can use it while building things
COPY deploy/infra/helm/helm-v* /usr/bin/helm
RUN chmod +x /usr/bin/helm

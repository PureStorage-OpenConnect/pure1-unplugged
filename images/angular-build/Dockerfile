# FROM debian:latest

# # replace shell with bash so we can source files
# RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# # update the repository sources list
# # and install dependencies
# RUN apt-get update \
#     && apt-get install -y curl \
#     && apt-get -y autoclean

# # nvm environment variables
# ENV NVM_DIR /ng-app/nvm
# ENV NODE_VERSION 10.5.0
# # install nvm
# # https://github.com/creationix/nvm#install-script
# RUN curl --silent -o- https://raw.githubusercontent.com/creationix/nvm/v0.31.2/install.sh | bash

# # install node and npm
# RUN source $NVM_DIR/nvm.sh \
#     && nvm install $NODE_VERSION \
#     && nvm alias default $NODE_VERSION \
#     && nvm use default

# # add node and npm to path so the commands are available
# ENV NODE_PATH $NVM_DIR/v$NODE_VERSION/lib/node_modules
# ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

# # confirm installation
# RUN node -v
# RUN npm -v
# # RUN npm config set prefix '/ng-app/gui/node_modules'
FROM node:10.3.0 as builder

RUN mkdir /.npm
RUN chmod 777 /.npm
RUN mkdir /.config
RUN chmod 777 /.config

# Fix for 'Error: could not get uid/gid' when running in Jenkins
RUN npm config set unsafe-perm true

RUN apt-get -y update
RUN apt-get -y upgrade
# Required to install chrome from repo
RUN apt-get install -y apt-transport-https
# Install make
RUN apt-get install -y build-essential

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] https://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
RUN apt-get -y update
# Install Chrome for testing
RUN apt-get install -y google-chrome-stable

WORKDIR /ng-app

# Add a stage of the dex image so that we can copy any necessary executables/files over
FROM quay.io/dexidp/dex:v2.12.0

FROM nginx:1.13.3-alpine

LABEL authors = "support.purestorage.com"

# Necessary step from dex dockerfile, adds certificates because some connectors need them
RUN apk add --update ca-certificates openssl

# Copy the dex executable and web content over
COPY --from=0 /usr/local/bin/dex /usr/local/bin/dex
COPY --from=0 /web /web
ADD images/pure1-unplugged/dex-web-content /web

ADD build/bin /usr/bin/

RUN mkdir -p /srv/pure1-unplugged/

ADD build/gui /srv/pure1-unplugged/
COPY images/pure1-unplugged/nginx/err_html /srv/pure1-unplugged/

COPY images/pure1-unplugged/nginx/nginx-default.conf /etc/nginx/conf.d/default.conf

ADD images/pure1-unplugged/kibana /kibana-files

RUN apk update && apk add curl

EXPOSE 8000
EXPOSE 80

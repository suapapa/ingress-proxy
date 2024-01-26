# build stage
FROM golang:alpine as builder

ARG PROGRAM_VER=dev-docker

## build gcsfuse v0.41.6 is latest release at point of writing (2022-09-11)
ARG GCSFUSE_VERSION=v0.41.6

RUN go install github.com/googlecloudplatform/gcsfuse@${GCSFUSE_VERSION}

# RUN apt-get -qq update && \
# 	apt-get install -yqq upx

COPY . /build
WORKDIR /build

RUN go build -ldflags "-X main.programVer=${PROGRAM_VER}" -o app
# RUN strip /build/app
# RUN upx -q -9 /build/app

# ---
FROM alpine:latest

RUN apk add --update --no-cache \
	fuse \
	certbot

# try to renew letsencrypt ssl cert in every 12 hours
RUN SLEEPTIME=$(awk 'BEGIN{srand(); print int(rand()*(3600+1))}'); \
	crontab -l | { cat; echo "0 0,12 * * * root sleep $SLEEPTIME && certbot renew -q"; } | \
	crontab -

COPY --from=builder /build/create_ssl_cert.sh /bin/create_ssl_cert.sh
RUN chmod +x /bin/create_ssl_cert.sh

COPY --from=builder /build/app /bin/app

## install gcsfuse
COPY --from=builder /go/bin/gcsfuse /usr/bin

ENV TELEGRAM_APITOKEN="secret"
ENV TELEGRAM_ROOM_ID="secret"

EXPOSE 9001
EXPOSE 443
EXPOSE 80

RUN mkdir /bucket
RUN ln -s /bucket/cert /etc/letsencrypt

WORKDIR /bin

CMD ['sh', '-c', 'crond && ./app']

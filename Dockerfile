## base image
FROM golang:1.12.8

MAINTAINER Craig Tomkow "ctomkow@gmail.com"

ARG UID
ARG GID
ARG NAME

# system app dependency
RUN apt-get update && apt-get install -y \
	mariadb-client

# app user
RUN groupadd -g ${GID} ${NAME} && useradd -s /bin/bash -u ${UID} -g ${GID} ${NAME}

RUN mkdir -p /go/src/github.com/ctomkow/tto
WORKDIR /go/src/github.com/ctomkow/tto

RUN go get "github.com/takama/daemon"       && \
    go get "github.com/golang/glog"         && \
    go get "github.com/robfig/cron"         && \
    go get "github.com/fsnotify/fsnotify"   && \
    go get "github.com/go-sql-driver/mysql" && \
    go get "golang.org/x/crypto/ssh"

COPY . /go/src/github.com/ctomkow/tto
RUN go install

# install app
RUN mkdir -p /etc/tto && \
    mkdir -p /opt/tto
RUN /go/bin/tto install

CMD ["/go/bin/tto"]

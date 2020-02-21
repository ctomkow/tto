# tto

[![Build Status](https://travis-ci.org/ctomkow/tto.svg?branch=master)](https://travis-ci.org/ctomkow/tto)

tto [t⋅toe]: _3-2-1 MySQL backup and sync_. Three backups, two copies on different storage, one located off-site.

→	An asynchronous client-server app for synchronizing a MySQL database between two systems. In addition, it keeps a ring buffer of _X_ backups on the secondary system. 

The main use-case for developing this was to help maintain a hybrid primary / [primary/secondary] application 
deployment where replication was not possible.

### Use Cases
* Replace cron scheduled database backup shell scripts
* Don't want to/can't setup MySQL replication
* Enable a simple primary/secondary infrastructure across two data centers


### Build Dependencies
* `"github.com/fsnotify/fsnotify"`
* `"github.com/golang/glog"`
* `"github.com/robfig/cron"`
* `"github.com/takama/daemon"`
* `"github.com/go-sql-driver/mysql"`
* `"golang.org/x/crypto/ssh"`

### Runtime Dependencies
* `mysqldump`
* `InnoDB tables`

# Install

The application needs to be installed on the primary and secondary systems. Each will be configured for their 
respective roles (sender | receiver).

## RPM Install

    sudo yum install tto-<version>.x86_64.rpm

(edit /etc/tto/conf.json)

    sudo systemctl start tto   

## RPM Uninstall
    sudo yum remove tto

WARNING: this removes working dir /opt/tto and conf dir /etc/tto as well!

## Docker Install
Currently, the docker install doesn't create a sample conf.json at runtime. See the sample conf.json included in the repo.

    mkdir /etc/tto

    mkdir /opt/tto

    cp conf.json /etc/tto/

(edit /etc/tto/conf.json)

`docker build --build-arg GID=` _myGID_ ` --build-arg UID=` _myUID_ ` --build-arg NAME=` _myUsername_ ` -t tto .`

    docker run -v /etc/tto/conf.json:/etc/tto/conf.json -v /opt/tto:/opt/tto tto

## Docker Compose

    mkdir /etc/tto
    mkdir /opt/tto
    cp conf.json /etc/tto/

(edit /etc/tto/conf.json)

(edit .env)

    docker-compose up -d

## Manual Build & Install
(`go get` all build dependencies)

    go build tto.go

    ./tto install

(edit /etc/tto/conf.json)

    systemctl start tto

## Manual Uninstall

    ./tto remove
    rm -r /opt/tto/
    rm -r /etc/tto/

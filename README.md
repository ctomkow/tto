# tto
3-2-1 MySQL backup and sync v0.1.2

An asynchronous client-server app for synchronizing a MySQL database between two systems. The
main use-case for developing this was to help maintain a hybrid primary / [primary / secondary] application 
deployment where replication was not possible.

### Use Cases
* Replace cron scheduled shell scripts
* Don't want to/can't setup MySQL replication
* Enable a simple primary/secondary infrastructure across two data centers


### Build Dependencies
* "github.com/fsnotify/fsnotify"
* "github.com/golang/glog"
* "github.com/robfig/cron"
* "github.com/takama/daemon"
* "github.com/go-sql-driver/mysql"
* "golang.org/x/crypto/ssh"

### Runtime Dependencies
* mysqldump
* InnoDB tables

# Install

The application needs to be installed on the primary and secondary systems. Each will be configured for their 
respective roles (sender | receiver).

## Docker Install
Currently, the docker install doesn't create a sample conf.json at runtime. See the sample conf.json included in the repo.

`mkdir /etc/tto`

`mkdir /opt/tto`

`cp conf.json /etc/tto/`

(edit /etc/tto/conf.json)

`docker build --build-arg GID=` **myGID** ` --build-arg UID=` **myUID** ` --build-arg NAME=` **myUsername** ` -t tto .`

`docker run -v /etc/tto/conf.json:/etc/tto/conf.json -v /opt/tto:/opt/tto tto`

## Docker Compose

`mkdir /etc/tto`

`mkdir /opt/tto`

`cp conf.json /etc/tto/`

(edit /etc/tto/conf.json)

(edit .env)

`docker-compose up -d`

## OS Install
(`go get` all build dependencies)

`go build tto.go`

`./tto install`

(edit /etc/tto/conf.json)

`systemctl start tto`

## Uninstall

`./tto remove`

`rm -r /opt/tto/`

`rm -r /etc/tto/`

# tto
3-2-1 MySQL backup and sync v0.1.1

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

## Docker Install

`docker-compose up -d`

## Install (on both systems)

`go build tto.go`

`sudo ./tto install`

(edit /etc/conf.json)

`sudo systemctl start tto`

## Uninstall

`sudo ./tto remove`

`rm -r /opt/tto/`

`rm -r /etc/tto/`

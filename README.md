# TFTP [![GoDoc](http://godoc.org/github.com/ikor/tftp?status.png)](http://godoc.org/github.com/ikor/tftp)
[![Build Status](https://travis-ci.org/ikor/tftp.svg?branch=master)](https://travis-ci.org/ikor/tftp)

TFTP server implementation with in memory storage. Current version implements [RFC 1350](https://tools.ietf.org/html/rfc1350)


## Build 
```sh
$ make build
```

## Run 

Server bind :1069 port. Port is configurable via TFTP_PORT variable
```sh
Server
$ ./cmd/server/ramtftp
```
Test with tftp client
```sh
$ tftp 127.0.0.1 1069
```
To change port number set TFT_PORT environment variable.
```sh
$ TFTP_PORT=10690 ./cmd/server/ramtftp
```
## 
```sh
$ make test
```

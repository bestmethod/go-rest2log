# go-rest2log
A Go service for forwarding rest API calls to go-logger. Useful for when you have a client that can't access any logging service.

## Usage
```
go get github.com/bestmethod/go-rest2log
cd ~/go/src/github.com/bestmethod/go-rest2log
go build rest2log.go
```

That's it :) For production systems, you only need the resulting rest2log binary and the config file. Don't forget to edit the config file to your liking.

This assumes that logging will happen via devlog or by means of journald/docker-logs. It gives a way to log into those using a rest interface for those systems that need a basic system like this one.

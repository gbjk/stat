export GOPATH := $(shell pwd)

all: bin/stats

bin/stats: src/github.com/nf/stats/stats.go
	go install github.com/nf/stats

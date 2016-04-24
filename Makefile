export GOPATH := $(shell pwd)

all: bin/stats

bin/stats: src/github.com/gbjk/stats/stats.go
	go install github.com/gbjk/stats

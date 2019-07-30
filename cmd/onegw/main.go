package main

import (

	"github.com/pmker/onegw"
	"github.com/pmker/onegw/cmd/plugins"
	_ "net/http/pprof"
)

// gvite is the official command-line client for Vite

func main() {
	onegw.PrintBuildVersion()
	plugins.Loading()
}

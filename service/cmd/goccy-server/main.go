// Command goccy-server is the /ingest server using github.com/goccy/go-json.
package main

import (
	"os"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/goccy"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server"
)

func main() { os.Exit(server.Run(goccy.Codec{}, os.Args[1:])) }

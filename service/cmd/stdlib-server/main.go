// Command stdlib-server is the /ingest server using encoding/json.
package main

import (
	"os"

	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/codec/stdlib"
	"github.com/lurodrisilva/golang-json-encoding-benchmark/service/internal/server"
)

func main() { os.Exit(server.Run(stdlib.Codec{}, os.Args[1:])) }

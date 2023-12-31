// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// protoc-gen-emmylua is a plugin for the Google protocol buffer compiler to generate
// Go code. Install it by building this program and making it accessible within
// your PATH with the name:
//
//	protoc-gen-emmylua
//
// The 'go' suffix becomes part of the argument for the protocol compiler,
// such that it can be invoked as:
//
//	protoc --emmylua_out=paths=source_relative:. path/to/file.proto
//
// This generates Go bindings for the protocol buffer defined by file.proto.
// With that input, the output will be written to:
//
//	path/to/file.lua
//
// See the README and documentation for protocol buffers to learn more:
//
//	https://developers.google.com/protocol-buffers/
package main

import (
	"flag"
	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"protoc-gen-emmylua/genlua"
)

func main() {
	var (
		flags flag.FlagSet
	)
	// Try Debug ,Sleep and attach
	//time.Sleep(10 * time.Second)
	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			genlua.GenerateFile(gen, f)
		}
		gen.SupportedFeatures = gengo.SupportedFeatures
		return nil
	})
}

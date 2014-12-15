// Copyright 2014 The Transporter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// seed is a reimagining of the seed mongo to mongo tool

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/transporter"
)

var (
	sourceUri     = flag.String("s", "mongodb://localhost/", "source mongo uri to connect to")
	destUri       = flag.String("d", "mongodb://localhost/", "destination mongo uri to write documents to")
	sourceNS      = flag.String("source-ns", "", "the source namespace to copy")
	destinationNS = flag.String("dest-ns", "", "the destination namespace")
	tail          = flag.Bool("o", false, "tail the oplog")
	debug         = flag.Bool("v", false, "debug, dumps all the documents to stdout")
)

func init() {
	flag.Parse()
}

func main() {
	source :=
		transporter.NewNode("source", "mongo", map[string]interface{}{"uri": *sourceUri, "namespace": *sourceNS, "tail": *tail}).
			Add(transporter.NewNode("out", "mongo", map[string]interface{}{"uri": *destUri, "namespace": *destinationNS}))

	if *debug {
		source.Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))
	}

	pipeline, err := transporter.NewPipeline(source, events.NewLogEmitter(), 1*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeline.Run()
}

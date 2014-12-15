// Copyright 2014 The Transporter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// cat-mongo builds a transporter pipeline that can be used to fetch all
// the documents from a mongo collection, and emit them to stdout

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
	uri       = flag.String("s", "mongodb://localhost/", "mongo uri to connect to")
	namespace = flag.String("ns", "", "the namespace to cat")
	tail      = flag.Bool("o", false, "tail the oplog")
)

func init() {
	flag.Parse()
}

func main() {
	source :=
		transporter.NewNode("source", "mongo", map[string]interface{}{"uri": *uri, "namespace": *namespace, "debug": false, "tail": *tail}).
			Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))

	pipeline, err := transporter.NewPipeline(source, events.NewNoopEmitter(), 1*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeline.Run()
}

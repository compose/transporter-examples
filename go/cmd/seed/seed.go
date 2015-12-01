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

	"gopkg.in/mgo.v2"

	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/transporter"
)

var (
	sourceUri     = flag.String("s", "mongodb://localhost/", "source mongo uri to connect to")
	destUri       = flag.String("d", "mongodb://localhost/", "destination mongo uri to write documents to")
	sourceDB      = flag.String("source-db", "", "the source namespace to copy")
	destinationDB = flag.String("dest-db", "", "the destination namespace")
	debug         = flag.Bool("v", false, "debug, dumps all the documents to stdout")
	bulk          = flag.Bool("bulk", false, "bulk insert into the destination mongo")
)

func init() {
	flag.Parse()
}

func main() {

	sess, err := mgo.Dial(*sourceUri)
	if err != nil {
		fmt.Println("Can't connect: " + err.Error())
		os.Exit(1)
	}
	sess.Close()

	if *sourceDB == "" || *destinationDB == "" {
		fmt.Fprintln(os.Stderr, "source and destination database must be provided, exiting")
		os.Exit(1)
	}

	srcNamespace := fmt.Sprintf("%s./.*/", *sourceDB)
	destNamespace := fmt.Sprintf("%s./.*/", *destinationDB)

	source :=
		transporter.NewNode(fmt.Sprintf("source-%s", *sourceDB), "mongo", map[string]interface{}{"uri": *sourceUri, "namespace": srcNamespace, "tail": false}).
			Add(transporter.NewNode(fmt.Sprintf("dest-%s", *destinationDB), "mongo", map[string]interface{}{"uri": *destUri, "namespace": destNamespace, "bulk": *bulk}))

	if *debug {
		source.Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))
	}

	pipeline, err := transporter.NewPipeline(source, events.NewJsonLogEmitter(), 2*time.Second, nil, 10*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	pipeline.Run()
	if pipeline.Err != nil {
		fmt.Fprintf(os.Stderr, "Pipeline Errored with %v\n", pipeline.Err)
		os.Exit(1)
	}
	fmt.Println("Complete")
}

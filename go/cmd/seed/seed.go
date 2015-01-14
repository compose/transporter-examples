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
	"strings"
	"sync"
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

	names, err := sess.DB(*sourceDB).CollectionNames()
	if err != nil {
		fmt.Println("Error: " + err.Error())
	}

	if *sourceDB == "" || *destinationDB == "" || len(names) == 0 {
		fmt.Fprintln(os.Stderr, "No collections to copy, exiting")
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	for _, name := range names {

		if strings.HasPrefix(name, "system.") {
			continue
		}

		srcNamespace := fmt.Sprintf("%s.%s", *sourceDB, name)
		destNamespace := fmt.Sprintf("%s.%s", *destinationDB, name)

		source :=
			transporter.NewNode("source", "mongo", map[string]interface{}{"uri": *sourceUri, "namespace": srcNamespace, "tail": false}).
				Add(transporter.NewNode("out", "mongo", map[string]interface{}{"uri": *destUri, "namespace": destNamespace}))

		if *debug {
			source.Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))
		}

		pipeline, err := transporter.NewPipeline(source, events.NewLogEmitter(), 1*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		wg.Add(1)
		go func() {
			pipeline.Run()
			if pipeline.Err != nil {
				fmt.Fprintf(os.Stderr, "Pipeline Errored with %v\n", pipeline.Err)
				os.Exit(1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("Complete")
}

// Copyright 2014 The Transporter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// seed is a reimagining of the seed mongo to mongo tool

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"

	_ "github.com/compose/transporter/pkg/adaptor/all"
	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/transporter"
)

var (
	sourceURI     = flag.String("s", "mongodb://localhost/", "source mongo uri to connect to")
	destURI       = flag.String("d", "mongodb://localhost/", "destination mongo uri to write documents to")
	sourceDB      = flag.String("source-db", "", "the source namespace to copy")
	destinationDB = flag.String("dest-db", "", "the destination namespace")
	debug         = flag.Bool("v", false, "debug, dumps all the documents to stdout")
	bulk          = flag.Bool("bulk", false, "bulk insert into the destination mongo")
)

func init() {
	flag.Parse()
}

func dPrintf(s string, args ...interface{}) {
	if *debug {
		fmt.Printf(s, args)
	}
}

func testConn(uri string, ssl bool) {
	dialInfo, err := mgo.ParseURL(uri)
	dialInfo.FailFast = true
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		if ssl {
			dPrintf("dialing ssl")
			tlsConfig := &tls.Config{}
			tlsConfig.InsecureSkipVerify = true
			return tls.Dial("tcp", addr.String(), tlsConfig)
		}
		dPrintf("dialing non-ssl %+v\n", dialInfo)
		return net.Dial("tcp", addr.String())
	}
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Println("Can't connect: " + err.Error())
		os.Exit(1)
	}
	dPrintf("success")
	sess.Close()
}

type logger struct{}

func (l *logger) Output(calldepth int, s string) error {
	fmt.Printf("[%d] %s\n", calldepth, s)
	return nil
}

func main() {
	mgo.SetDebug(*debug)
	if *debug {
		mgo.SetLogger(&logger{})
	}
	sourceSSL := strings.Contains(*sourceURI, "ssl=true")
	if sourceSSL {
		*sourceURI = strings.Replace(*sourceURI, "ssl=true", "", -1)
	}
	destSSL := strings.Contains(*destURI, "ssl=true")
	if destSSL {
		*destURI = strings.Replace(*destURI, "ssl=true", "", -1)
	}
	dPrintf("testing source %s - ssl: %v...\n", *sourceURI, sourceSSL)
	testConn(*sourceURI, sourceSSL)
	dPrintf("testing dest %s - ssl: %v...\n", *destURI, destSSL)
	testConn(*destURI, destSSL)

	if *sourceDB == "" || *destinationDB == "" {
		fmt.Fprintln(os.Stderr, "source and destination database must be provided, exiting")
		os.Exit(1)
	}

	sourceNamespace := fmt.Sprintf("%s./.*/", *sourceDB)
	destNamespace := fmt.Sprintf("%s./.*/", *destinationDB)
	sourceConfig := map[string]interface{}{"uri": *sourceURI, "namespace": sourceNamespace, "tail": false}
	if sourceSSL {
		sourceConfig["ssl"] = map[string]interface{}{"cacerts": []string{}} // empty cacerts forces insecure no verify
	}
	destConfig := map[string]interface{}{"uri": *destURI, "namespace": destNamespace, "bulk": *bulk}
	if destSSL {
		destConfig["ssl"] = map[string]interface{}{"cacerts": []string{}} // empty cacerts forces insecure no verify
	}
	dPrintf("creating transport")
	source :=
		transporter.NewNode(fmt.Sprintf("source-%s", *sourceDB), "mongodb", sourceConfig).
			Add(transporter.NewNode(fmt.Sprintf("dest-%s", *destinationDB), "mongodb", destConfig))

	if *debug {
		source.Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))
	}
	pipeline, err := transporter.NewPipeline(source, events.NewJSONLogEmitter(), 2*time.Second, nil, 10*time.Second)
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

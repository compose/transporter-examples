// Copyright 2014 The Transporter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// ticker demonstrates the writing of a custom database adapter

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/compose/transporter/pkg/adaptor"
	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/message"
	"github.com/compose/transporter/pkg/pipe"
	"github.com/compose/transporter/pkg/transporter"
	"gopkg.in/mgo.v2/bson"
)

var (
	interval = flag.String("i", "500ms", "mongo uri to connect to")
)

func init() {
	flag.Parse()
}

func main() {
	// register our custom adaptor so that it's available to ransporter
	adaptor.Register("ticker", NewTicker)

	// construct the nodes
	source :=
		transporter.NewNode("source", "ticker", adaptor.Config{"interval": *interval}).
			Add(transporter.NewNode("out", "file", adaptor.Config{"uri": "stdout://"}))

	pipeline, err := transporter.NewPipeline(source, events.NewNoopEmitter(), 1*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeline.Run()
}

// Ticker is an example transporter adaptor, that just uses a time.Ticker
// to emit messages
type Ticker struct {
	ticker *time.Ticker
	pipe   *pipe.Pipe
	path   string
}

// NewTicker instantiates the adaptor, the structure here is very important, the adaptor.Registry requires a
//   func(pipe.Pipe, adaptor.ExtraConfig) (adaptor.StopStartListner, error)
func NewTicker(p *pipe.Pipe, path string, extra adaptor.Config) (adaptor.StopStartListener, error) {
	interval, err := time.ParseDuration(extra.GetString("interval"))
	if err != nil {
		return nil, err
	}
	return &Ticker{pipe: p, ticker: time.NewTicker(interval), path: path}, nil
}

// Start starts the adaptor as a source
func (t *Ticker) Start() error {
	for tm := range t.ticker.C {
		doc := bson.M{"id": bson.NewObjectId(), "timestamp": tm.UnixNano()}
		msg := message.NewMsg(message.Insert, doc)
		t.pipe.Send(msg)
	}
	return nil
}

// Stop stops the ticker
func (t *Ticker) Stop() error {
	t.ticker.Stop()
	return nil
}

// the adaptor.StopStartListener interface requires a Listen method, but in this
// case we're not doing anything, so lets just return an error
func (t *Ticker) Listen() error {
	return fmt.Errorf("Ticker can't listen")
}

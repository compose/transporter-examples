// Copyright 2014 The Transporter Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// ticker demonstrates the writing of a custom database adapter

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/compose/transporter-examples/go/cmd/twitter/tstreamer"
	"github.com/compose/transporter/pkg/adaptor"
	"github.com/compose/transporter/pkg/events"
	"github.com/compose/transporter/pkg/message"
	"github.com/compose/transporter/pkg/pipe"
	"github.com/compose/transporter/pkg/transporter"
	"gopkg.in/mgo.v2/bson"
)

func init() {
	flag.Parse()
}

func readConf() (string, string, string, string) {
	f, _ := os.Open("twitter.conf")
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Scan()
	consumerKey := s.Text()
	s.Scan()
	consumerSecret := s.Text()
	s.Scan()
	accessToken := s.Text()
	s.Scan()
	accessTokenSecret := s.Text()
	return consumerKey, consumerSecret, accessToken, accessTokenSecret
}

var (
	destURI       = flag.String("d", "mongodb://localhost/", "destination mongo uri to write documents to")
	destinationNS = flag.String("dest-ns", "", "the destination namespace")
	debug         = flag.Bool("v", false, "debug, dumps all the documents to stdout")
)

func main() {
	// register our custom adaptor so that it's available to Transporter
	adaptor.Register("twitter", NewTwitter)

	ck, cs, at, ats := readConf()

	// construct the nodes
	source :=
		transporter.NewNode("source", "twitter", map[string]interface{}{
			"consumerkey":       ck,
			"consumersecret":    cs,
			"accesstoken":       at,
			"accesstokensecret": ats}).
			Add(transporter.NewNode("out", "mongo", map[string]interface{}{"uri": *destURI, "namespace": *destinationNS}))

	if *debug {
		source.Add(transporter.NewNode("out", "file", map[string]interface{}{"uri": "stdout://"}))
	}

	pipeline, err := transporter.NewPipeline(source, events.NewNoopEmitter(), 1*time.Second)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeline.Run()
}

// Twitter is an example transporter adaptor, that just uses a time.Ticker
// to emit messages
type Twitter struct {
	timeline *tstreamer.TimelineListen
	feed     <-chan tstreamer.Tweet
	pipe     *pipe.Pipe
	path     string
}

// TwitterConfig provides the configuration options for a Twitter adaptor
type TwitterConfig struct {
	ConsumerKey       string `json:"consumerkey"`
	ConsumerSecret    string `json:"consumersecret"`
	AccessToken       string `json:"accesstoken"`
	AccessTokenSecret string `json:"accesstokensecret"`
}

// NewTwitter instantiates the adaptor, the structure here is very important, the adaptor.Registry requires a
//   func(pipe.Pipe, adaptor.ExtraConfig) (adaptor.StopStartListner, error)
func NewTwitter(p *pipe.Pipe, path string, extra adaptor.Config) (adaptor.StopStartListener, error) {
	var (
		conf TwitterConfig
		err  error
	)

	if err = extra.Construct(&conf); err != nil {
		return nil, err
	}

	if conf.ConsumerKey == "" || conf.ConsumerSecret == "" {
		return nil, fmt.Errorf("Both consumerkey and consumersecret required")
	}
	if conf.AccessToken == "" || conf.AccessTokenSecret == "" {
		return nil, fmt.Errorf("Both accesstoken and accesstokensecret required")
	}

	timeline, err := tstreamer.New(
		"https://stream.twitter.com/1.1/statuses/sample.json",
		conf.ConsumerKey,
		conf.ConsumerSecret,
		conf.AccessToken,
		conf.AccessTokenSecret,
	)

	if err != nil {
		return nil, err
	}

	feed := timeline.Listen()

	return &Twitter{
		timeline: timeline,
		feed:     feed,
		pipe:     p,
		path:     path}, nil
}

// Start starts the adaptor as a source
func (t *Twitter) Start() error {
	for tw := range t.feed {
		doc := bson.M{"id": bson.NewObjectId(), "tweet": tw}
		msg := message.NewMsg(message.Insert, doc)
		t.pipe.Send(msg)
	}
	return nil
}

// Stop stops the ticker
func (t *Twitter) Stop() error {
	// t.ticker.Stop()
	return nil
}

// Listen - the adaptor.StopStartListener interface requires a Listen method, but in this
// case we're not doing anything, so lets just return an error
func (t *Twitter) Listen() error {
	return fmt.Errorf("Ticker can't listen")
}

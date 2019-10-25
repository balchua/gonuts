// Copyright 2016-2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

var usageStr = `
Usage: stan-pub [options] <subject> <message>
 
Options:
	-s,  --server   <url>                    NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>           NATS Streaming cluster name
	-cr, --creds    <credentials>            NATS 2.0 Credentials
	-d, --delay     <delay in milliseconds>  Delay in publishing message in milliseonds.
`

// NOTE: Use tls scheme for TLS, e.g. stan-pub -s tls://demo.nats.io:4443 foo hello
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func main() {
	var (
		clusterID string
		clientID  string
		URL       string
		userCreds string
		delay     int
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "local-stan", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "local-stan", "The NATS Streaming cluster ID")
	flag.StringVar(&userCreds, "cr", "", "Credentials File")
	flag.StringVar(&userCreds, "creds", "", "Credentials File")
	flag.IntVar(&delay, "d", 1000, "Delay in seconds between publishing message")
	flag.IntVar(&delay, "delay", 1000, "Delay in seconds between publishing message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("Go Nuts Publisher")}
	// Use UserCredentials
	if userCreds != "" {
		opts = append(opts, nats.UserCredentials(userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	clientID = strconv.FormatInt(time.Now().UnixNano(), 10)
	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	defer sc.Close()

	subj := args[0]

	for {
		t := time.Now()
		msg := []byte("Message is : " + t.String())
		err = sc.Publish(subj, msg)
		if err != nil {
			log.Fatalf("Error during publish: %v\n", err)
		}
		log.Printf("Published [%s] : '%s'\n", subj, msg)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

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
	"os/signal"
	"strconv"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
)

var usageStr = `
Usage: stan-sub [options] <subject>

Options:
	-s,  --server   <url>            NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>   NATS Streaming cluster name
	-cr, --creds    <credentials>    NATS 2.0 Credentials

Subscription Options:
	--qgroup <name>                  Queue group
	--durable <name>                 Durable subscriber name
	--delay <in milliseconds>        Delay in milliseconds between consumption
`

// NOTE: Use tls scheme for TLS, e.g. stan-sub -s tls://demo.nats.io:4443 foo
func usage() {
	log.Fatalf(usageStr)
}

func printMsg(m *stan.Msg, i int) {
	//	if !m.Redelivered {

	log.Printf("[#%d] Received: %s\n", i, m)
	time.Sleep(100 * time.Millisecond)
	if err := m.Ack(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var (
		clusterID, clientID string
		URL                 string
		userCreds           string
		showTime            bool
		qgroup              string
		durable             string
		delay               int
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "local-stan", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "local-stan", "The NATS Streaming cluster ID")
	flag.BoolVar(&showTime, "t", false, "Display timestamps")
	// Subscription options
	flag.StringVar(&durable, "durable", "", "Durable subscriber name")
	flag.StringVar(&qgroup, "qgroup", "", "Queue group name")
	flag.StringVar(&userCreds, "cr", "", "Credentials File")
	flag.StringVar(&userCreds, "creds", "", "Credentials File")
	flag.IntVar(&delay, "d", 1000, "Delay in seconds between publishing message")
	flag.IntVar(&delay, "delay", 1000, "Delay in seconds between publishing message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Printf("Error: A subject must be specified.")
		usage()
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("Go Nuts Subscriber")}
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

	log.Printf("Client ID is %s", clientID)
	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", URL, clusterID, clientID)

	// Process Subscriber Options.
	startOpt := stan.StartAt(pb.StartPosition_NewOnly)

	subj, i := args[0], 0
	mcb := func(msg *stan.Msg) {
		i++
		time.Sleep(time.Duration(delay) * time.Millisecond)
		printMsg(msg, i)
	}

	_, err = sc.QueueSubscribe(subj, qgroup, mcb, startOpt, stan.DurableName(durable), stan.SetManualAckMode(), stan.MaxInflight(1))
	if err != nil {
		sc.Close()
		log.Fatal(err)
	}

	log.Printf("Listening on [%s], clientID=[%s], qgroup=[%s] durable=[%s]\n", subj, clientID, qgroup, durable)

	if showTime {
		log.SetFlags(log.LstdFlags)
	}

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			sc.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

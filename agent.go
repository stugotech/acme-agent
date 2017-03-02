package main

import (
	"os"
	"sort"

	"strings"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

const (
	listenFlag         = "listen"
	listenDefault      = ":8080"
	logFlag            = "log"
	logDefault         = "info"
	pathPrefixFlag     = "pathPrexix"
	pathPrefixDefault  = ".well-known/acme-challenge"
	storeFlag          = "store"
	storeDefault       = "etcd"
	storeNodesFlag     = "store-nodes"
	storeNodesDefault  = "127.0.0.1:2379"
	storePrefixFlag    = "store-prefix"
	storePrefixDefault = "acme-agent"
)

func main() {
	app := cli.NewApp()
	app.Name = "acme-agent"
	app.Usage = "solve ACME HTTP challenges using a KV store"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  storeFlag,
			Usage: "key value store to use [etcd|consul|boltdb|zookeeper]",
			Value: storeDefault,
		},
		cli.StringFlag{
			Name:  storeNodesFlag,
			Usage: "comma-separated list of KV nodes (authority only)",
			Value: storeNodesDefault,
		},
		cli.StringFlag{
			Name:  storePrefixFlag,
			Usage: "prefix in KV store",
			Value: storePrefixDefault,
		},
		cli.StringFlag{
			Name:  logFlag,
			Usage: "logging level [debug|info|warning|error|fatal|panic]",
			Value: logDefault,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "serve",
			Usage:  "Run a HTTP server to respond to challenges",
			Action: serve,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  listenFlag,
					Value: listenDefault,
					Usage: "TCP interface or unix socket to listen on",
				},
				cli.StringFlag{
					Name:  pathPrefixFlag,
					Value: pathPrefixDefault,
					Usage: "TCP interface or unix socket to listen on",
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	log.WithFields(log.Fields{"pid": os.Getpid()}).Info("**** LOGGING STARTED ****")
	server, err := NewServer(createServerConfig(c))
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("error creating server")
		return fmt.Errorf("serve: error creating server: %v", err)
	}
	err = server.Listen()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("fatal error")
		return fmt.Errorf("serve: unhandled error: %v", err)
	}
	return nil
}

func createConfig(c *cli.Context) *Config {
	level, err := log.ParseLevel(c.GlobalString(logFlag))

	if err != nil {
		log.Warnf("invalid log level %q, defaulting to \"info\"", level)
		level = log.InfoLevel
	}

	log.SetLevel(level)

	return &Config{
		Store:       c.GlobalString(storeFlag),
		StoreNodes:  strings.Split(c.GlobalString(storeNodesFlag), ","),
		StorePrefix: c.GlobalString(storePrefixFlag),
	}
}

func createServerConfig(c *cli.Context) *ServerConfig {
	return &ServerConfig{
		Config:     *createConfig(c),
		Listen:     c.String(listenFlag),
		PathPrefix: c.String(pathPrefixFlag),
	}
}

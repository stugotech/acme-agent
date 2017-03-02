package main

import (
	"fmt"
	"path"

	"net/http"

	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
)

const (
	serverErrorFormat        = "server: unhandled error: %v"
	storeCreationErrorFormat = "server: can't create store '%s': %v"
	storeGetErrorFormat      = "server: can't get key '%s': %v"
	storeDeleteErrorFormat   = "server: can't delete key '%s': %v"
)

const (
	pathRegexp = "^/%s/([a-zA-Z0-9_-]+)$"
)

type serverInfo struct {
	config    *ServerConfig
	store     store.Store
	validPath *regexp.Regexp
}

type serverInfoHandler func(s *serverInfo, response http.ResponseWriter, request *http.Request)

// ServerConfig describes the configuration of the server
type ServerConfig struct {
	Config
	Listen     string
	PathPrefix string
}

// Server describes an ACME challenge server
type Server interface {
	Listen() error
}

// NewServer creates a new server
func NewServer(config *ServerConfig) (Server, error) {
	etcd.Register()
	consul.Register()
	boltdb.Register()
	zookeeper.Register()

	log.WithFields(log.Fields{
		"Store":       config.Store,
		"StoreNodes":  config.StoreNodes,
		"StorePrefix": config.StorePrefix,
		"Listen":      config.Listen,
		"PathPrefix":  config.PathPrefix,
	}).Info("creating new server")

	storeConfig := &store.Config{}
	s, err := libkv.NewStore(store.Backend(config.Store), config.StoreNodes, storeConfig)

	if err != nil {
		log.WithFields(log.Fields{
			"err":   err,
			"store": config.Store,
		}).Errorf("server: can't create store")
		return nil, fmt.Errorf(storeCreationErrorFormat, config.Store, err)
	}

	config.PathPrefix = strings.Trim(config.PathPrefix, "/")
	validPath := regexp.MustCompile(fmt.Sprintf(pathRegexp, config.PathPrefix))

	return &serverInfo{
		config:    config,
		store:     s,
		validPath: validPath,
	}, nil
}

// Listen starts the server listening for connections
func (s *serverInfo) Listen() error {
	http.HandleFunc("/", s.makeHandler(challengeHandler))
	log.Infof("listening on %q", s.config.Listen)
	err := http.ListenAndServe(s.config.Listen, nil)

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("server: encountered error")
		return fmt.Errorf(serverErrorFormat, err)
	}

	return nil
}

func challengeHandler(s *serverInfo, response http.ResponseWriter, request *http.Request) {
	match := s.validPath.FindStringSubmatch(request.URL.Path)
	if match == nil {
		log.WithFields(log.Fields{"path": request.URL.Path}).Error("server: invalid URL format")
		http.NotFound(response, request)
		return
	}

	key := match[1]
	value, err := s.getValue(key)

	if err != nil {
		log.WithFields(log.Fields{"key": key, "err": err}).Error("server: error retrieving key")
		http.NotFound(response, request)
		return
	}

	response.Write(value)

	if err = s.deleteKey(key); err != nil {
		log.WithFields(log.Fields{"err": err, "key": key}).Error("server: error deleting key")
	}
}

func (s *serverInfo) makeHandler(fn serverInfoHandler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		fn(s, response, request)
	}
}

func (s *serverInfo) getValue(key string) ([]byte, error) {
	key = path.Join(s.config.StorePrefix, key)
	log.Debugf("looking up key %q from KV store", key)
	kv, err := s.store.Get(key)

	if err != nil {
		return nil, fmt.Errorf(storeGetErrorFormat, key, err)
	}

	return kv.Value, nil
}

func (s *serverInfo) deleteKey(key string) error {
	key = path.Join(s.config.StorePrefix, key)
	err := s.store.Delete(key)

	if err != nil {
		return fmt.Errorf(storeDeleteErrorFormat, key, err)
	}

	return nil
}

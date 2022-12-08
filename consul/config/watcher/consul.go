package consul

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/hashicorp/consul/api"
)

// Option is etcd config option.
type Option func(o *options)

type options struct {
	ctx  context.Context
	path string
}

// WithContext with registry context.
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithPath is config path
func WithPath(p string) Option {
	return func(o *options) {
		o.path = p
	}
}

type Source interface {
	Load() ([]*KeyValue, error)
	Watch() (Watcher, error)
}

// Watcher watches a source for changes.
type Watcher interface {
	Next() ([]*KeyValue, error)
	Stop() error
}

type KeyValue struct {
	Key    string
	Value  []byte
	Format string
}


type source struct {
	client  *api.Client
	options *options
}

func New(client *api.Client, opts ...Option) (Source, error) {
	options := &options{
		ctx:  context.Background(),
		path: "",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.path == "" {
		return nil, errors.New("path invalid")
	}

	return &source{
		client:  client,
		options: options,
	}, nil
}

// Load return the config values
func (s *source) Load() ([]*KeyValue, error) {
	kv, _, err := s.client.KV().List(s.options.path, nil)
	if err != nil {
		return nil, err
	}

	pathPrefix := s.options.path
	if !strings.HasSuffix(s.options.path, "/") {
		pathPrefix = pathPrefix + "/"
	}
	kvs := make([]*KeyValue, 0)
	for _, item := range kv {
		k := strings.TrimPrefix(item.Key, pathPrefix)
		if k == "" {
			continue
		}
		kvs = append(kvs, &KeyValue{
			Key:    k,
			Value:  item.Value,
			Format: strings.TrimPrefix(filepath.Ext(k), "."),
		})
	}
	return kvs, nil
}

// Watch return the watcher
func (s *source) Watch() (Watcher, error) {
	return newWatcher(s)
}

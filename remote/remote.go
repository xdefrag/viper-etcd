package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/spf13/viper"
	etcd "go.etcd.io/etcd/client"
)

func (p provider) Get(rp viper.RemoteProvider) (io.Reader, error) {
	return get(rp)
}

func (p provider) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	return get(rp)
}

func (p provider) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	watcher, err := watcher(rp)
	if err != nil {
		return nil, nil
	}

	rr := make(chan *viper.RemoteResponse)
	done := make(chan bool)

	ctx, cancel := context.WithCancel(ctx())

	go func(done <-chan bool) {
		select {
		case <-done:
			cancel()
		}
	}(done)

	go func(ctx context.Context, rr chan<- *viper.RemoteResponse, key string) {
		for {
			res, err := watcher.Next(ctx)

			if err == context.Canceled {
				return
			}

			if err != nil {
				rr <- &viper.RemoteResponse{
					Error: err,
				}

				continue
			}

			rr <- &viper.RemoteResponse{
				Value: readr(rp, res.Node),
			}
		}

	}(ctx, rr, rp.Path())

	return rr, done
}

type provider struct{}

func init() {
	viper.RemoteConfig = &provider{}
}

func newEtcdClient(rp viper.RemoteProvider) (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{
		Endpoints: []string{rp.Endpoint()},
	})
	if err != nil {
		return nil, err
	}

	return etcd.NewKeysAPI(client), nil
}

func ctx() context.Context {
	return context.Background()
}

func get(rp viper.RemoteProvider) (io.Reader, error) {
	kapi, err := newEtcdClient(rp)
	if err != nil {
		return nil, err
	}

	res, err := kapi.Get(ctx(), rp.Path(), &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(readr(rp, res.Node)), nil
}

func watcher(rp viper.RemoteProvider) (etcd.Watcher, error) {
	kapi, err := newEtcdClient(rp)
	if err != nil {
		return nil, err
	}

	return kapi.Watcher(rp.Path(), &etcd.WatcherOptions{
		Recursive: true,
	}), nil
}

func readr(rp viper.RemoteProvider, node *etcd.Node) []byte {
	vars := make(map[string]string)
	nodeWalk(rp, node, vars)

	b, _ := json.Marshal(vars)

	return b
}

func nodeWalk(rp viper.RemoteProvider, node *etcd.Node, vars map[string]string) {
	if node != nil {
		k := node.Key
		if !node.Dir {
			if rp.SecretKeyring() != "" {
				// TODO decode
			}

			vars[keyReplace(k, rp.Path())] = node.Value
		} else {
			for _, node := range node.Nodes {
				nodeWalk(rp, node, vars)
			}
		}
	}
}

func keyReplace(key, pre string) string {
	return strings.ReplaceAll(strings.TrimPrefix(key, pre+"/"), "/", ".")
}

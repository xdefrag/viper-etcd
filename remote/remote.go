package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	go func(ctx context.Context, rr chan<- *viper.RemoteResponse) {
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

	}(ctx, rr)

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
	b, _ := json.Marshal(nodeVals(node, rp.Path(), rp.SecretKeyring()).(map[string]interface{}))

	return b
}

func nodeVals(node *etcd.Node, path, keyring string) interface{} {
	if !node.Dir && path == node.Key {
		if keyring != "" {
			//
		}

		return node.Value
	}

	vv := make(map[string]interface{})

	if len(node.Nodes) == 0 {
		newKey := keyFirstChild(strings.ReplaceAll(node.Key, path, ""))
		vv[newKey] = nodeVals(node, fmt.Sprintf("%s/%s", path, newKey), keyring)

		return vv
	}

	for _, n := range node.Nodes {
		vv[keyLastChild(n.Key)] = nodeVals(n, n.Key, keyring)
	}

	return vv
}

func keyLastChild(key string) string {
	kk := strings.Split(key, "/")

	return kk[len(kk)-1]
}

func keyFirstChild(key string) string {
	kk := strings.Split(key, "/")

	return kk[1]
}

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

type Decoder interface {
	Decode(io.Reader) (interface{}, error)
}

type Config struct {
	Decoder
	viper.RemoteProvider

	Username string
	Password string
}

func (c *Config) Get(rp viper.RemoteProvider) (io.Reader, error) {
	c.verify(rp)
	c.RemoteProvider = rp

	return c.get()
}

func (c *Config) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	c.verify(rp)
	c.RemoteProvider = rp

	return c.get()
}

func (c *Config) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {
	c.verify(rp)
	c.RemoteProvider = rp

	watcher, err := c.watcher()
	if err != nil {
		return nil, nil
	}

	rr := make(chan *viper.RemoteResponse)
	done := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())

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
				Value: c.readr(res.Node),
			}
		}

	}(ctx, rr)

	return rr, done
}

func (c Config) verify(rp viper.RemoteProvider) {
	if rp.Provider() != "etcd" {
		panic("Viper-etcd remote supports only etcd.")
	}

	if rp.SecretKeyring() != "" {
		panic("Viper-etcd doesn't support keyrings, use Decoder instead.")
	}
}

func (c Config) newEtcdClient() (etcd.KeysAPI, error) {
	client, err := etcd.New(etcd.Config{
		Username: c.Username,
		Password: c.Password,

		Endpoints: []string{c.Endpoint()},
	})
	if err != nil {
		return nil, err
	}

	return etcd.NewKeysAPI(client), nil
}

func (c Config) get() (io.Reader, error) {
	kapi, err := c.newEtcdClient()
	if err != nil {
		return nil, err
	}

	res, err := kapi.Get(context.Background(), c.Path(), &etcd.GetOptions{
		Recursive: true,
	})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(c.readr(res.Node)), nil
}

func (c Config) watcher() (etcd.Watcher, error) {
	kapi, err := c.newEtcdClient()
	if err != nil {
		return nil, err
	}

	return kapi.Watcher(c.Path(), &etcd.WatcherOptions{
		Recursive: true,
	}), nil
}

func (c Config) readr(node *etcd.Node) []byte {
	m, _ := json.Marshal(c.nodeVals(node, c.Path()).(map[string]interface{}))

	return m
}

func (c Config) nodeVals(node *etcd.Node, path string) interface{} {
	if !node.Dir && path == node.Key {

		if c.Decoder != nil {
			val, _ := c.Decode(strings.NewReader(node.Value))

			return val
		}

		return node.Value
	}

	vv := make(map[string]interface{})

	if len(node.Nodes) == 0 {
		newKey := keyFirstChild(strDiff(node.Key, path))
		vv[newKey] = c.nodeVals(node, pathSum(path, newKey))

		return vv
	}

	for _, n := range node.Nodes {
		vv[keyLastChild(n.Key)] = c.nodeVals(n, n.Key)
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

func strDiff(s1, s2 string) string {
	return strings.ReplaceAll(s1, s2, "")
}

func pathSum(pp ...string) string {
	return strings.Join(pp, "/")
}

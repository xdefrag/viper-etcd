package main

import (
	"context"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
	_ "github.com/xdefrag/viper-etcd/remote"
	etcd "go.etcd.io/etcd/client"
)

// You can start ETCD container like this:
// docker run -d -p 2379:2379 quay.io/coreos/etcd:latest etcd --advertise-client-urls http://0.0.0.0:2380 --listen-client-urls http://0.0.0.0:2379

func main() {
	if os.Getenv("ETCD_ADDR") == "" {
		os.Setenv("ETCD_ADDR", "http://0.0.0.0:2379")
	}

	initEtcdKeys()

	vpr := viper.New()

	must(vpr.AddRemoteProvider("etcd", os.Getenv("ETCD_ADDR"), "/testconfig"))

	vpr.SetConfigType("json")

	must(vpr.ReadRemoteConfig())

	go func() {
		for {
			must(vpr.WatchRemoteConfig())
		}
	}()

	for {
		spew.Dump(vpr.AllSettings())
		time.Sleep(2 * time.Second)
	}
}

func initEtcdKeys() {
	client, err := etcd.New(etcd.Config{
		Endpoints: []string{os.Getenv("ETCD_ADDR")},
	})
	if err != nil {
		panic(err)
	}

	kapi := etcd.NewKeysAPI(client)

	ctx := context.Background()

	must2(kapi.Set(ctx, "/testconfig/access/token", "some_token", nil))
	must2(kapi.Set(ctx, "/testconfig/redis/addr", "http://0.0.0.0:6379", nil))
	must2(kapi.Set(ctx, "/testconfig/redis/password", "veryStrongPassword", nil))
	must2(kapi.Set(ctx, "/testconfig/deeply/nested/config/wow", "this_is_value", nil))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2(_ interface{}, err error) {
	if err != nil {
		panic(err)
	}
}

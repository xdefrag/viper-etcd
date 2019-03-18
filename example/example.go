package main

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/xdefrag/viper-etcd/remote"
	etcd "go.etcd.io/etcd/client"
)

func init() {
	viper.RemoteConfig = &remote.Config{
		Decoder: &decode{},
	}
}

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

type decode struct{}

func (d decode) Decode(r io.Reader) (interface{}, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	s := string(raw)

	if strings.Contains(s, ",") {
		return strings.Split(s, ","), nil
	}

	return s, nil
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

	must2(kapi.Set(ctx, "/testconfig/access/token", mustEncode("some_token"), nil))
	must2(kapi.Set(ctx, "/testconfig/redis/addr", mustEncode("http://0.0.0.0:6379"), nil))
	must2(kapi.Set(ctx, "/testconfig/redis/password", mustEncode("veryStrongPassword"), nil))
	must2(kapi.Set(ctx, "/testconfig/deeply/nested/config/wow", mustEncode("this_is_value"), nil))
	must2(kapi.Set(ctx, "/testconfig/providers", mustEncode([]string{"redis", "postgres"}), nil))
	must2(kapi.Set(ctx, "/testconfig/lucky/numbers", mustEncode([]int{9, 13}), nil))
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

func mustEncode(raw interface{}) string {
	var val string
	switch raw := raw.(type) {
	case []string:
		val = strings.Join(raw, ",")
	case []int:
		var ss []string
		for _, r := range raw {
			ss = append(ss, strconv.Itoa(r))
		}
		val = strings.Join(ss, ",")
	default:
		val = cast.ToString(raw)
	}

	return val
}

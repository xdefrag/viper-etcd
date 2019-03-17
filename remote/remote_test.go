package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jeremywohl/flatten"
	etcd "go.etcd.io/etcd/client"
)

func TestRemote(t *testing.T) {
	drp := newDefaultRemoteProvider()

	if drp.Endpoint() == "http://" {
		t.Skip("Tests require ETCD_ADDR")
	}

	etcdKeysInit(t)
	defer etcdKeysClear(t)

	t.Run("Get", func(t *testing.T) {
		r, err := newProvider().Get(drp)
		if err != nil {
			t.Error(err)
		}

		assertMap(t, r, testconfig)
	})

	t.Run("Watch", func(t *testing.T) {
		r, err := newProvider().Watch(drp)
		if err != nil {
			t.Error(err)
		}

		assertMap(t, r, testconfig)
	})

	t.Run("WatchChannel", func(t *testing.T) {
		rr, done := newProvider().WatchChannel(drp)
		time.Sleep(time.Second)

		c, err := newEtcdClient(newDefaultRemoteProvider())
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.Set(context.Background(), "/testconfig/access/token", "newtoken", nil)
		if err != nil {
			t.Fatal(err)
		}

		r := <-rr

		if r.Error != nil {
			t.Fatal(err)
		}

		assertMap(t, bytes.NewReader(r.Value), map[string]interface{}{
			"access": map[string]interface{}{
				"token": "newtoken",
			},
		})

		done <- false
	})
}

func assertMap(t *testing.T, r io.Reader, m map[string]interface{}) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	var tc map[string]interface{}
	if err := json.Unmarshal(b, &tc); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tc, m) {
		t.Errorf("Got: %v, want: %v", tc, m)
	}
}

var testconfig = map[string]interface{}{
	"database": map[string]interface{}{
		"addr":     "http://localhost:5432",
		"password": "testing_password",
		"username": "testing_username",
	},
	"access": map[string]interface{}{
		"token": "testing_token",
	},
}

func etcdKeysInit(t *testing.T) {
	kapi, err := newEtcdClient(newDefaultRemoteProvider())
	if err != nil {
		t.Fatal(err)
	}

	tc, err := flatten.Flatten(testconfig, "", flatten.DotStyle)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range tc {
		k = "/testconfig/" + strings.Replace(k, ".", "/", -1)
		must2(t)(kapi.Set(context.Background(), k, v.(string), nil))
	}
}

func etcdKeysClear(t *testing.T) {
	kapi, err := newEtcdClient(newDefaultRemoteProvider())
	if err != nil {
		t.Fatal(err)
	}

	must2(t)(kapi.Delete(context.Background(), "/testconfig", &etcd.DeleteOptions{
		Recursive: true,
	}))
}

func newDefaultRemoteProvider() defaultRemoteProvider {
	return defaultRemoteProvider{
		provider:      "etcd",
		endpoint:      "http://" + os.Getenv("ETCD_ADDR"),
		path:          "/testconfig",
		secretKeyring: "",
	}
}

func newProvider() *provider {
	return &provider{}
}

type defaultRemoteProvider struct {
	provider      string
	endpoint      string
	path          string
	secretKeyring string
}

func (rp defaultRemoteProvider) Provider() string {
	return rp.provider
}

func (rp defaultRemoteProvider) Endpoint() string {
	return rp.endpoint
}

func (rp defaultRemoteProvider) Path() string {
	return rp.path
}

func (rp defaultRemoteProvider) SecretKeyring() string {
	return rp.secretKeyring
}

func must2(t *testing.T) func(interface{}, error) {
	return func(_ interface{}, err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
}

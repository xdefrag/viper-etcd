# Viper Etcd
[![Build Status](https://travis-ci.com/xdefrag/viper-etcd.svg?branch=master)](https://travis-ci.com/xdefrag/viper-etcd)
[![codecov](https://codecov.io/gh/xdefrag/viper-etcd/branch/master/graph/badge.svg)](https://codecov.io/gh/xdefrag/viper-etcd)
[![Go Report Card](https://goreportcard.com/badge/github.com/xdefrag/viper-etcd)](https://goreportcard.com/report/github.com/xdefrag/viper-etcd)

Advanced ETCD remote config provider for [Viper](https://github.com/spf13/viper). Default remote config needs json on specific key, this one recursively get all nested key-values.

**Attention!** This remote provider doesn't support SecretKeyring. [Crypt](https://github.com/xordataexchange/crypt) that Viper used for encryption is [abandoned](https://github.com/xordataexchange/crypt/issues/23) and [doesn't work well with latest GPG versions](https://github.com/xordataexchange/crypt/issues/12). BUT you can use custom decoder.

# Usage

Instead of requiring Viper's default remote package, initialize RemoteConfig by yourself like this:
```go
import (
        "github.com/xdefrag/viper-etcd/remote"
	"github.com/spf13/viper"
)

func init() {
        viper.RemoteConfig = &remote.Config{
                Username: "user",
                Password: "pass", // etcd credentials if any

                Decoder: &decode{}, // struct with remote.Decoder interface: 
                // type Decoder interface {
                //         Decode(io.Reader) (interface{}, error)
                // }
                // Set your encryption or serialization tool here.
        }
}
```
And that's it! More details in [example](https://github.com/xdefrag/viper-etcd/blob/master/example/example.go).

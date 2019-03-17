# Viper Etcd
[![Build Status](https://travis-ci.com/xdefrag/viper-etcd.svg?branch=master)](https://travis-ci.com/xdefrag/viper-etcd)
[![codecov](https://codecov.io/gh/xdefrag/viper-etcd/branch/master/graph/badge.svg)](https://codecov.io/gh/xdefrag/viper-etcd)
[![Go Report Card](https://goreportcard.com/badge/github.com/xdefrag/viper-etcd)](https://goreportcard.com/report/github.com/xdefrag/viper-etcd)

Advanced ETCD remote config provider for [Viper](https://github.com/spf13/viper). Default remote config needs json on specific key, this one recursively get all nested key-values.

**Attention!** This remote provider doesn't support SecretKeyring. [Crypt](https://github.com/xordataexchange/crypt) that Viper used for encryption is [abandoned](https://github.com/xordataexchange/crypt/issues/23) and [doesn't work well with latest GPG versions](https://github.com/xordataexchange/crypt/issues/12).

In your viper configurations do this:
```go
import (
        // _ "github.com/spf13/viper/remote"
        _ "github.com/xdefrag/viper-etcd/remote"
)
```
And that's it. More details in [example](https://github.com/xdefrag/viper-etcd/blob/master/example/example.go).

# Viper Etcd
[![Build Status](https://travis-ci.com/xdefrag/viper-etcd.svg?branch=master)](https://travis-ci.com/xdefrag/viper-etcd)
[![codecov](https://codecov.io/gh/xdefrag/viper-etcd/branch/master/graph/badge.svg)](https://codecov.io/gh/xdefrag/viper-etcd)

Advanced ETCD remote config provider for viper.

*Attention!* This remote provider doesn't support SecretKeyring.

In your viper configurations do this:
```go
import (
        // _ "github.com/spf13/viper/remote"
        _ "github.com/xdefrag/viper-etcd/remote"
)
```

And it will work. I hope. Anyway it work in progress, but no more ugly jsons, with this package you can enjoy real deal etcd key value experience!

More details soon.

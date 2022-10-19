package naming

import (
	_ "github.com/istepheny/naming/registry/consul"
	_ "github.com/istepheny/naming/registry/etcd"
	_ "github.com/istepheny/naming/registry/zookeeper"
)

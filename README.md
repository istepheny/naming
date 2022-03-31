# Naming

## 安装

 ```shell
 go get git.ucloudadmin.com/monkey/naming@latest
 ```

## 使用

```go
package main

import (
	"fmt"

	"git.ucloudadmin.com/monkey/naming"
	"git.ucloudadmin.com/monkey/naming/app"
	"git.ucloudadmin.com/monkey/naming/config"
)

func main() {
	// 配置
	c := config.Config{
		Driver:   "consul",                   // 可选项 consul etcd zookeeper
		Servers:  []string{"127.0.0.1:8500"}, // 服务端地址
		Username: "",                         // 用户名
		Password: "my-token",                 // 密码，如果 Driver 是 consul，填写 consul 的 Token
	}

	// 初始化
	n := naming.New(c)

	// 服务注册
	n.Register(app.New("my-app", "127.0.0.1", 80))

	// 服务注册附带自定义数据
	metadata := map[string]string{"version": "0.0.1"}
	n.Register(app.New("my-app", "127.0.0.1", 80, metadata))

	// 服务发现
	myApp, err := n.Discover("my-app")
	fmt.Println(myApp.Address, err) // 127.0.0.1:80

	// 获取服务的所有节点
	myApps, err := n.DiscoverAll("my-app")
	for _, myApp := range myApps {
		fmt.Println(myApp.Address, err)
	}

	// 服务注销
	n.Deregister(app.New("my-app", "127.0.0.1", 80))
}
```

### 参考

- https://github.com/go-kiss/sniper
- https://github.com/go-kit/kit
- https://github.com/go-kratos/kratos
- https://github.com/gogf/katyusha
- https://github.com/yoyofx/yoyogo
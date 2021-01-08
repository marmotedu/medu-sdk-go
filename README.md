# medu-sdk-go介绍

marmotedu另一种sdk实现思路（仿照：tencentcloud-sdk-go、aws-sdk-go、ucloud-sdk-go、jdcloud-sdk-go等）。
medu-sdk-go提供iam-authz-server组件/v1/authz接口的go版SDK。

使用方法为：

```go
package main

import (
	"fmt"

	"github.com/marmotedu/medu-sdk-go/sdk"
	"github.com/marmotedu/medu-sdk-go/services/iam"
	"github.com/ory/ladon"
)

func main() {
	client, _ := iam.NewClientWithSecret("XhbY3aCrfjdYcP1OFJRu9xcno8JzSbUIvGE2", "bfJRvlFwsoW9L30DlG87BBW0arJamSeK")

	req := iam.NewAuthzRequest()
	req.Resource = sdk.String("resources:articles:ladon-introduction")
	req.Action = sdk.String("delete")
	req.Subject = sdk.String("users:peter")
	ctx := ladon.Context(map[string]interface{}{"remoteIP": "192.168.0.5"})
	req.Context = &ctx

	resp, err := client.Authz(req)
	if err != nil {
		fmt.Println("err1", err)
		return
	}
	fmt.Printf("get response body: `%s`\n", resp.String())
	fmt.Printf("allowed: %v\n", resp.Allowed)
}
```

输出如下：


```bash
get response body: `{"code":0,"allowed":true}`
allowed: true
```


## 参考

- 代码请参考：[authz.go](./examples/authz.go)
- 其它marmotedu SDK实现：[marmotedu-sdk-go](https://github.com/marmotedu/marmotedu-sdk-go)

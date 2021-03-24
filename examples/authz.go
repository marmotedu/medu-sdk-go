package main

import (
	"fmt"

	"github.com/ory/ladon"

	"github.com/marmotedu/medu-sdk-go/sdk"
	iam "github.com/marmotedu/medu-sdk-go/services/iam/authz"
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

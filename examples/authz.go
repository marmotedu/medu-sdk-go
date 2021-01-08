package main

import (
	"fmt"

	"github.com/ory/ladon"

	"github.com/marmotedu/medu-sdk-go/sdk"
	iamv1 "github.com/marmotedu/medu-sdk-go/services/iam/v1"
)

func main() {
	client, _ := iamv1.NewClientWithSecret("XhbY3aCrfjdYcP1OFJRu9xcno8JzSbUIvGE2", "bfJRvlFwsoW9L30DlG87BBW0arJamSeK")

	req := iamv1.NewAuthzRequest()
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

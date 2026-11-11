package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func apiKeysExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Create API Key
	params := &sendio.CreateApiKeyRequest{
		Name: "nice api key",
	}

	resp, err := client.ApiKeys.CreateWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created API Key id: " + resp.Id)
	fmt.Println("Token: " + resp.Token)

	// List
	apiKeys, err := client.ApiKeys.ListWithContext(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("You have %d api keys in your project\n", len(apiKeys.Data))

	// Update
	updateParams := &sendio.UpdateApiKeyRequest{
		Name: "nicer api key",
	}
	updateResp, err := client.ApiKeys.UpdateWithContext(ctx, resp.Id, updateParams)
	if err != nil {
		panic(err)
	}
	fmt.Println("Updated API Key id: " + updateResp.Id)

	// Delete
	_, err = client.ApiKeys.RemoveWithContext(ctx, resp.Id)
	if err != nil {
		panic(err)
	}
	println("deleted api key id: " + resp.Id)
}

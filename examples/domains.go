package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/send-io/sendio-go"
)

func domainsExample() {
	ctx := context.TODO()
	apiKey := os.Getenv("SENDIO_API_KEY")

	client := sendio.NewClient(apiKey)

	// Create Domain params
	params := &sendio.CreateDomainRequest{
		Name:             "drish.dev",
		Region:           "us-east-1",
		CustomReturnPath: "outbound",
		Capabilities: &sendio.DomainCapabilities{
			Sending:   sendio.DomainCapabilityStatusEnabled,
			Receiving: sendio.DomainCapabilityStatusEnabled,
		},
	}

	domain, err := client.Domains.CreateWithContext(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created Domain entry id: " + domain.Id)
	fmt.Println("Status: " + domain.Status)
	if domain.Capabilities != nil {
		fmt.Println("Sending: " + domain.Capabilities.Sending)
		fmt.Println("Receiving: " + domain.Capabilities.Receiving)
	}

	for _, record := range domain.Records {
		fmt.Printf("%v\n", record)
	}

	// Get
	retrievedDomain, err := client.Domains.GetWithContext(ctx, domain.Id)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Retrieved domain: %v", retrievedDomain)
	if retrievedDomain.Capabilities != nil {
		fmt.Println("Sending: " + retrievedDomain.Capabilities.Sending)
		fmt.Println("Receiving: " + retrievedDomain.Capabilities.Receiving)
	}

	updateDomainParams := &sendio.UpdateDomainRequest{
		OpenTracking:  true,
		ClickTracking: true,
		Tls:           sendio.Enforced,
	}

	updated, err := client.Domains.UpdateWithContext(ctx, domain.Id, updateDomainParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", updated)

	// List
	domains, err := client.Domains.ListWithContext(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("You have %d domains in your project\n", len(domains.Data))

	// Verify
	verified, err := client.Domains.VerifyWithContext(ctx, domain.Id)
	if err != nil {
		panic(err)
	}
	if verified {
		println("verified domain id: " + domain.Id)
	} else {
		println("could not verify domain id: " + domain.Id)
	}

	// Remove
	removed, err := client.Domains.RemoveWithContext(ctx, domain.Id)
	if err != nil {
		panic(err)
	}
	if removed {
		println("removed domain id: " + domain.Id)
	} else {
		println("could not remove domain id: " + domain.Id)
	}
}

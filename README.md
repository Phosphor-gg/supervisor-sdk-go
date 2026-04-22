# Supervisor Go SDK

Official Go SDK for the [Supervisor](https://supervisor.gg) content moderation API.

Zero dependencies — uses only the standard library.

## Installation

```bash
go get github.com/Phosphor-gg/supervisor-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    supervisor "github.com/Phosphor-gg/supervisor-sdk-go"
)

func main() {
    client := supervisor.NewClient("sk-...")

    result, err := client.Moderate(context.Background(), &supervisor.ModerationRequest{
        Text: supervisor.String("check this text"),
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Flagged: %v\n", result.Flagged)
    fmt.Printf("Labels: %v\n", result.Labels)
}
```

## Usage

### Moderate Text

```go
result, err := client.Moderate(ctx, &supervisor.ModerationRequest{
    Text:  supervisor.String("some text to check"),
    Model: &supervisor.ModelSentinel,
})
```

### Batch Moderation

```go
results, err := client.ModerateBatch(ctx, &supervisor.BatchModerationRequest{
    Texts: []string{"first", "second", "third"},
})
for _, r := range results {
    fmt.Printf("Flagged: %v, Labels: %v\n", r.Flagged, r.Labels)
}
```

### Username Check

```go
result, err := client.CheckUsername(ctx, "username123")
fmt.Printf("Flagged: %v, Score: %f\n", result.Flagged, result.Score)
```

### Get Labels

```go
labels, err := client.GetLabels(ctx)
```

## Platform API

```go
platform := supervisor.NewPlatformClient("client-id", "client-secret")

// Provision a user
user, err := platform.ProvisionUser(ctx, "user@example.com")

// Moderate on behalf of a user
result, err := platform.Moderate(ctx, &supervisor.PlatformModerationRequest{
    UserEmail: "user@example.com",
    Text:      supervisor.String("check this"),
})

// List linked users
users, err := platform.ListUsers(ctx)

// Create checkout
checkout, err := platform.CreateCheckout(ctx, &supervisor.PlatformCheckoutRequest{
    UserEmail:    "user@example.com",
    Tier:         supervisor.TierStandard,
    BillingCycle: supervisor.BillingMonthly,
    SuccessURL:   "https://yourapp.com/success",
    CancelURL:    "https://yourapp.com/cancel",
})
```

## Configuration

```go
client := supervisor.NewClient("sk-...",
    supervisor.WithBaseURL("https://api.supervisor.gg"),
    supervisor.WithTimeout(30 * time.Second),
    supervisor.WithHTTPClient(customClient),
)
```

## Error Handling

```go
result, err := client.Moderate(ctx, req)
if err != nil {
    var apiErr *supervisor.Error
    if errors.As(err, &apiErr) {
        if apiErr.IsAuthError() {
            log.Fatal("Invalid API key")
        }
        if apiErr.IsRateLimit() {
            log.Fatal("Rate limited")
        }
        log.Fatalf("API error [%d]: %s", apiErr.StatusCode, apiErr.Message)
    }
    log.Fatal(err)
}
```

## License

MIT

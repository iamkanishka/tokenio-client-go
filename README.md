# tokenio-client-go

A production-grade Go client SDK for the [Token.io Open Banking platform](https://reference.token.io).

**Full API coverage** — every endpoint from `reference.token.io`, correctly modelled from the real OpenAPI specification.

---

## APIs Covered

| Package         | Endpoints                                                                      |
| --------------- | ------------------------------------------------------------------------------ |
| `payments`      | Initiate payment, Get payment(s), Embedded auth, QR code                       |
| `vrp`           | Create/Get/List/Revoke consents, Initiate/Get/List VRP payments, Confirm funds |
| `accountonfile` | Create/Get/Delete tokenized accounts                                           |
| `tokenrequests` | Store/Get token request, Get result, Initiate bank auth                        |
| `transfers`     | Redeem transfer token, Get/List transfers                                      |
| `tokens`        | Get/List tokens, Cancel token                                                  |
| `refunds`       | Initiate/Get/List refunds, Get by transfer                                     |
| `payouts`       | Initiate/Get/List payouts                                                      |
| `settlement`    | Create/Get accounts, Get transactions, CRUD rules, Get payouts                 |
| `ais`           | Get/List accounts, balances, standing orders, transactions                     |
| `banks`         | Get banks v1, Get bank countries, Get banks v2                                 |
| `subtpps`       | Create/Get/List/Delete sub-TPPs, Get children                                  |
| `authkeys`      | Submit/Get/Delete member keys                                                  |
| `reports`       | Get bank status(es)                                                            |
| `webhooks`      | Set/Get/Delete webhook config + typed event parsing                            |
| `verification`  | Initiate account verification check                                            |

---

## Requirements

Go 1.25+

## Installation

```bash
go get github.com/iamkanishka/tokenio-client-go
```

## Quick Start

```go
import (
    tokenio "github.com/iamkanishka/tokenio-client-go"
    "github.com/iamkanishka/tokenio-client-go/pkg/payments"
    "github.com/iamkanishka/tokenio-client-go/pkg/vrp"
    "github.com/iamkanishka/tokenio-client-go/pkg/common"
)

client, err := tokenio.NewClient(tokenio.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    // Environment: tokenio.EnvironmentProduction  // defaults to sandbox
})
if err != nil { log.Fatal(err) }
defer client.Close()
```

## Payments v2

```go
// Initiate a payment
payment, err := client.Payments.InitiatePayment(ctx, payments.InitiatePaymentRequest{
    Initiation: payments.PaymentInitiation{
        BankID: "ob-modelo",
        Amount: &payments.Amount{Value: "10.50", Currency: "GBP"},
        Creditor: &payments.PartyAccount{
            AccountNumber: "12345678",
            SortCode:      "040004",
            Name:          "Acme Ltd",
        },
        RemittanceInformationPrimary: "Invoice INV-001",
        CallbackURL:                  "https://yourapp.com/callback",
        FlowType:                     payments.FlowTypeFullHostedPages,
    },
    PispConsentAccepted: true,
})

// Handle redirect auth
if payment.Status.RequiresRedirect() {
    // redirect PSU to:
    redirectURL := payment.GetRedirectURL()
}

// Handle embedded auth
if payment.Status.RequiresEmbeddedAuth() {
    updated, err := client.Payments.ProvideEmbeddedAuth(ctx, payments.ProvideEmbeddedAuthRequest{
        PaymentID:    payment.ID,
        EmbeddedAuth: map[string]string{"otp_field_id": "123456"},
    })
}

// Poll to completion (prefer webhooks in production)
final, err := client.Payments.PollUntilFinal(ctx, payment.ID, payments.PollOptions{
    Interval: 2 * time.Second,
})

// Generate QR code
svgBytes, err := client.Payments.GenerateQRCode(ctx, payments.GenerateQRCodeRequest{
    Data: payment.GetRedirectURL(),
})
```

## Variable Recurring Payments (VRP)

```go
// Create consent
consent, err := client.VRP.CreateVrpConsent(ctx, vrp.CreateVrpConsentRequest{
    Initiation: vrp.VrpConsentInitiation{
        BankID:                  "ob-modelo",
        Currency:                "GBP",
        Creditor:                &common.PartyAccount{Name: "Acme", SortCode: "040004", AccountNumber: "12345678"},
        MaximumIndividualAmount: "500.00",
        PeriodicLimits: []vrp.PeriodicLimit{
            {MaximumAmount: "1000.00", PeriodType: "MONTH", PeriodAlignment: "CALENDAR"},
        },
        CallbackURL: "https://yourapp.com/vrp-callback",
    },
})

// Check funds availability
available, err := client.VRP.ConfirmFunds(ctx, consent.ID, "49.99")

// Initiate a VRP payment
payment, err := client.VRP.CreateVrp(ctx, vrp.CreateVrpRequest{
    Initiation: vrp.VrpInitiation{
        ConsentID: consent.ID,
        Amount:    &common.Amount{Value: "49.99", Currency: "GBP"},
    },
})
```

## AIS — Account Information Services

```go
accounts, err := client.AIS.GetAccounts(ctx, 50, "")
balance,  err := client.AIS.GetBalance(ctx, accounts.Accounts[0].ID)
txns,     err := client.AIS.GetTransactions(ctx, accountID, 20, "")
sos,      err := client.AIS.GetStandingOrders(ctx, 20, "")
```

## Refunds

```go
refund, err := client.Refunds.InitiateRefund(ctx, refunds.InitiateRefundRequest{
    RefundInitiation: refunds.RefundInitiation{
        TransferID: "t:abc",
        Amount:     &common.Amount{Value: "5.00", Currency: "GBP"},
    },
})
```

## Webhooks

```go
// Configure your webhook endpoint
client.Webhooks.SetConfig(ctx, webhooks.SetWebhookConfigRequest{
    Config: webhooks.WebhookConfig{URL: "https://yourapp.com/webhooks/token"},
})

// Parse and verify incoming events
event, err := client.Webhooks.Parse(body, r.Header.Get("X-Token-Signature"))
if event.Type == webhooks.EventTypePaymentCompleted {
    data, _ := webhooks.DecodePaymentEvent(event)
    fmt.Println(data.PaymentID, data.Status)
}
```

## Configuration Options

```go
client, err := tokenio.NewClient(
    tokenio.Config{ClientID: "...", ClientSecret: "..."},
    tokenio.WithEnvironment(tokenio.EnvironmentProduction),
    tokenio.WithTimeout(15 * time.Second),
    tokenio.WithMaxRetries(5),
    tokenio.WithRetryWait(500*time.Millisecond, 10*time.Second),
    tokenio.WithRateLimit(50, 10),
    tokenio.WithTracing(),                          // OpenTelemetry
    tokenio.WithLogger(zapLogger),                  // structured logging
    tokenio.WithWebhookSecret("your-wh-secret"),
)
```

## Error Handling

```go
import sdkerrors "github.com/iamkanishka/tokenio-client-go/internal/errors"

_, err := client.Payments.GetPayment(ctx, id)
if sdkerrors.IsNotFound(err)    { /* 404 */ }
if sdkerrors.IsUnauthorized(err){ /* 401 */ }
if sdkerrors.IsRateLimit(err)   { /* 429 */ }
if sdkerrors.IsRetryable(err)   { /* 429/5xx */ }
if sdkerrors.IsServerError(err) { /* 5xx */ }

var ae *sdkerrors.APIError
if errors.As(err, &ae) {
    fmt.Println(ae.Code, ae.Status, ae.Message, ae.RequestID)
}
```

## Running Tests

```bash
# Unit tests (no network)
go test ./tests/unit/... -v

# Integration tests (requires sandbox credentials)
TOKEN_CLIENT_ID=xxx TOKEN_CLIENT_SECRET=yyy \
  go test -tags=integration ./tests/integration/... -v
```

## Payment Status Reference

| Status                              | Final   | Requires Action                |
| ----------------------------------- | ------- | ------------------------------ |
| `INITIATION_PENDING`                | No      | Poll or wait for webhook       |
| `INITIATION_PENDING_REDIRECT_AUTH`  | No      | Redirect PSU                   |
| `INITIATION_PENDING_REDIRECT_HP`    | No      | Redirect PSU to Hosted Pages   |
| `INITIATION_PENDING_REDIRECT_PBL`   | No      | Redirect PSU (Pay By Link)     |
| `INITIATION_PENDING_EMBEDDED_AUTH`  | No      | Collect and submit auth fields |
| `INITIATION_PENDING_DECOUPLED_AUTH` | No      | Wait — bank contacts PSU       |
| `INITIATION_PROCESSING`             | No      | Wait                           |
| `INITIATION_COMPLETED`              | **Yes** | —                              |
| `INITIATION_REJECTED`               | **Yes** | —                              |
| `INITIATION_FAILED`                 | **Yes** | —                              |
| `INITIATION_DECLINED`               | **Yes** | —                              |
| `INITIATION_EXPIRED`                | **Yes** | —                              |
| `SETTLEMENT_IN_PROGRESS`            | No      | Wait                           |
| `SETTLEMENT_COMPLETED`              | **Yes** | —                              |
| `CANCELED`                          | **Yes** | —                              |

---

## License

MIT

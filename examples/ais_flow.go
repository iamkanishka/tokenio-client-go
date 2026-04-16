//go:build ignore

// ais_flow.go demonstrates the Token.io Account Information Services flow:
// list accounts → fetch balances → fetch transactions → fetch standing orders.
package main

import (
	"context"
	"fmt"
	"log"

	tokenio "github.com/iamkanishka/tokenio-client-go"
)

func main() {
	client, err := tokenio.NewClient(tokenio.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// ── 1. List accounts ──────────────────────────────────────────────────────
	accounts, err := client.AIS.GetAccounts(ctx, 50, "")
	if err != nil {
		log.Fatalf("GetAccounts: %v", err)
	}
	fmt.Printf("Found %d account(s)\n", len(accounts.Accounts))

	for _, acc := range accounts.Accounts {
		fmt.Printf("\nAccount: id=%s type=%s currency=%s\n",
			acc.ID, acc.Type, acc.Currency)

		// ── 2. Fetch balance ───────────────────────────────────────────────────
		balance, err := client.AIS.GetBalance(ctx, acc.ID)
		if err != nil {
			log.Printf("  GetBalance(%s): %v", acc.ID, err)
			continue
		}
		if balance.Current != nil {
			fmt.Printf("  Current balance: %s %s\n",
				balance.Current.Value, balance.Current.Currency)
		}
		if balance.Available != nil {
			fmt.Printf("  Available:       %s %s\n",
				balance.Available.Value, balance.Available.Currency)
		}

		// ── 3. Fetch recent transactions ───────────────────────────────────────
		txns, err := client.AIS.GetTransactions(ctx, acc.ID, 10, "")
		if err != nil {
			log.Printf("  GetTransactions(%s): %v", acc.ID, err)
			continue
		}
		fmt.Printf("  Transactions: %d\n", len(txns.Transactions))
		for _, tx := range txns.Transactions {
			amt := ""
			if tx.Amount != nil {
				amt = tx.Amount.Value + " " + tx.Amount.Currency
			}
			fmt.Printf("    [%s] %s %s\n", tx.Type, amt, tx.Description)
		}
	}

	// ── 4. List all balances at once ──────────────────────────────────────────
	balances, err := client.AIS.GetBalances(ctx, 50, "")
	if err != nil {
		log.Fatalf("GetBalances: %v", err)
	}
	fmt.Printf("\nAll balances: %d account(s)\n", len(balances.Balances))

	// ── 5. List standing orders ───────────────────────────────────────────────
	sos, err := client.AIS.GetStandingOrders(ctx, 20, "")
	if err != nil {
		log.Printf("GetStandingOrders (non-fatal): %v", err)
	} else {
		fmt.Printf("Standing orders: %d\n", len(sos.StandingOrders))
	}
}

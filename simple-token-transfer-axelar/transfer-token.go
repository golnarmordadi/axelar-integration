package main

import (
	"context"
	"fmt"
	"log"

	axelarsdk "github.com/axelarnetwork/axelar-go-sdk"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	// Set up Axelar client
	client, err := axelarsdk.NewClient(context.Background(), axelarsdk.Options{
		NodeAddress: "https://rpc.testnet.axelar.dev", // Use mainnet address for production
		ChainID:     "axelar-testnet",
	})
	if err != nil {
		log.Fatalf("Failed to create Axelar client: %v", err)
	}
	log.Println("Connected to Axelar testnet")

	// Transfer tokens
	err = transferTokens(client)
	if err != nil {
		log.Fatalf("Token transfer failed: %v", err)
	}
	log.Println("Token transfer successful!")
}

func transferTokens(client *axelarsdk.Client) error {
	// Set sender and recipient details
	senderAddress := "your-chain-sender-address"          // Your chain address
	recipientAddress := "ethereum:0xRecipientAddressHere" // Axelar-wrapped token destination
	amount := sdk.NewInt(1000000)                         // Amount in base units (e.g., 1 USDC = 1,000,000 micro-USDC)
	token := "uusdc"                                      // Token denomination (e.g., Axelar-supported USDC)

	// Build transfer message
	msg := client.TransferMessage{
		Sender:    senderAddress,
		Recipient: recipientAddress,
		Amount:    amount,
		Token:     token,
		Channel:   "channel-0", // Replace with your chain's channel ID to Axelar
	}

	// Submit transfer
	txResponse, err := client.Transfer(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed to send transfer: %w", err)
	}

	log.Printf("Transfer submitted. TxHash: %s", txResponse.TxHash)
	return nil
}

// Key Points
// 1- Middleware Required: Axelar handles the complexity of token wrapping, unwrapping, and cross-chain transfers.
// 2- Simple Setup: Once the IBC channel and relayer are running, you can use Axelar’s Go SDK to initiate and monitor transfers.
// 3- Security: Axelar’s relayers and validators ensure secure and reliable cross-chain transfers.

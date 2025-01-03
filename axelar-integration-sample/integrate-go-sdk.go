package main

import (
	"context"
	"log"

	axelarsdk "github.com/axelarnetwork/axelar-go-sdk"
	axelar "github.com/axelarnetwork/axelar-go-sdk/client"
	"github.com/axelarnetwork/axelar-go-sdk/client/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	// Connect to Axelar Client
	client, err := axelarsdk.NewClient(context.Background(), axelarsdk.Options{
		NodeAddress: "https://rpc.testnet.axelar.dev:443", // Use Axelar's testnet RPC endpoint
		ChainID:     "axelar-testnet",
	})
	if err != nil {
		log.Fatalf("Failed to connect to Axelar client: %v", err)
	}
	log.Println("Connected to Axelar testnet")

	// Example 1: Query Axelar Network
	queryAxelarNetwork(client)

	// Example 2: Send Tokens Cross-Chain
	sendTokens(client)

	// Example 3: General Message Passing
	generalMessagePassing(client)
}

// Example 1: Query Axelar Network
func queryAxelarNetwork(client *axelar.Client) {
	log.Println("Querying Axelar network...")

	// Query supported chains
	chains, err := client.EVMChains(context.Background())
	if err != nil {
		log.Fatalf("Failed to fetch EVM chains: %v", err)
	}
	log.Println("Supported EVM Chains:")
	for _, chain := range chains {
		log.Printf(" - Chain Name: %s, Chain ID: %d", chain.Name, chain.ChainID)
	}

	// Query token details
	tokenDetails, err := client.AssetDetails(context.Background(), "uusdc")
	if err != nil {
		log.Fatalf("Failed to fetch token details: %v", err)
	}
	log.Printf("Token Details: Name=%s, Decimals=%d", tokenDetails.Asset, tokenDetails.Decimals)
}

// Example 2: Send Tokens Cross-Chain
func sendTokens(client *axelar.Client) {
	log.Println("Sending tokens cross-chain...")

	// Sender and recipient details
	senderAddress := "axelar1...."                       // Replace with the sender's Axelar address
	recipientAddress := "ethereum:0xRecipientAddress..." // Replace with the recipient's Ethereum address
	amount := sdk.NewInt(1000000)                        // Amount to transfer (e.g., 1 USDC in micro-units)

	// Create a transfer request
	transferRequest := &types.Transfer{
		Recipient: recipientAddress,
		Amount:    amount,
		Asset:     "uusdc", // Replace with the correct asset identifier
	}

	// Execute the transfer
	txResponse, err := client.Transfer(context.Background(), senderAddress, transferRequest)
	if err != nil {
		log.Fatalf("Failed to send tokens: %v", err)
	}

	log.Printf("Transaction submitted successfully: TxHash=%s", txResponse.TxHash)
}

// Example 3: General Message Passing
func generalMessagePassing(client *axelar.Client) {
	log.Println("Sending general message...")

	// Create a general message
	generalMessage := types.GeneralMessage{
		SourceChain:   "cosmoshub",               // Source chain
		Destination:   "ethereum",                // Destination chain
		Payload:       []byte("Hello Ethereum!"), // Custom payload
		RecipientAddr: "0xRecipientAddress...",   // Recipient address on the destination chain
	}

	// Submit the message
	txResponse, err := client.SubmitGeneralMessage(context.Background(), generalMessage)
	if err != nil {
		log.Fatalf("Failed to send general message: %v", err)
	}

	log.Printf("General message submitted successfully: TxHash=%s", txResponse.TxHash)
}

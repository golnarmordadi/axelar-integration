// This code is part of an IBC middleware layer designed to:

// Enable custom logic for handling messages sent across chains.
// Properly manage token denominations during transfers.
// Enhance the flexibility of inter-chain communication within the Cosmos ecosystem.

package gmp_middleware

import (
	"fmt"

	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// GeneralMessageHandler defines the interface for handling general messages with or without tokens.
type GeneralMessageHandler interface {
	HandleGeneralMessage(ctx sdk.Context, srcChain, srcAddress string, destAddress string, payload []byte) error
	HandleGeneralMessageWithToken(ctx sdk.Context, srcChain, srcAddress string, destAddress string, payload []byte, coin sdk.Coin) error
}

// TODO: Replace this placeholder with the actual Axelar GMP account address.
const AxelarGMPAcc = ""

// Message represents a general message attached in the ICS20 packet memo field.
type Message struct {
	SourceChain   string `json:"source_chain"`
	SourceAddress string `json:"source_address"`
	Payload       []byte `json:"payload"`
	Type          int64  `json:"type"`
}

// MessageType enumerates the types of messages handled by this middleware.
type MessageType int

const (
	TypeUnrecognized            = iota // Unrecognized message type
	TypeGeneralMessage                 // General message
	TypeGeneralMessageWithToken        // General message that includes token information
)

// parseDenom converts the denomination to the receiver chain representation.
// It handles whether the receiver chain is the source of the token and removes or adds appropriate prefixes.
func parseDenom(packet channeltypes.Packet, denom string) string {
	log.Printf("Parsing denomination: %s for packet sequence: %d", denom, packet.Sequence)

	// If the receiver chain is the source of the token
	if types.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), denom) {
		log.Println("Receiver chain is the source of the token")

		// Remove the prefix added by the sender chain
		voucherPrefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		if len(denom) <= len(voucherPrefix) {
			log.Printf("Invalid denomination: %s. Cannot remove prefix as it is shorter than the prefix length.", denom)
			return denom
		}
		unprefixedDenom := denom[len(voucherPrefix):]

		// Coin denomination used for sending from the escrow address
		denom = unprefixedDenom

		// If the denomination is non-native, derive the IBC denomination
		denomTrace := types.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
			log.Printf("Parsed non-native denomination to IBC format: %s", denom)
		}

		return denom
	}

	// If the receiver chain is NOT the source of the token
	log.Println("Receiver chain is NOT the source of the token")

	prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + denom
	denom = transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()

	log.Printf("Parsed denomination with prefix added by destination chain: %s", denom)
	return denom
}

// validateMessage ensures that the Message struct is properly formed.
func validateMessage(msg *Message) error {
	if msg == nil {
		log.Println("Received a nil message for validation")
		return fmt.Errorf("message cannot be nil")
	}

	if msg.SourceChain == "" {
		log.Println("Validation failed: SourceChain is empty")
		return fmt.Errorf("source_chain field is required")
	}

	if msg.SourceAddress == "" {
		log.Println("Validation failed: SourceAddress is empty")
		return fmt.Errorf("source_address field is required")
	}

	if len(msg.Payload) == 0 {
		log.Println("Validation failed: Payload is empty")
		return fmt.Errorf("payload field is required")
	}

	if msg.Type != TypeGeneralMessage && msg.Type != TypeGeneralMessageWithToken {
		log.Printf("Validation failed: Unrecognized message type %d", msg.Type)
		return fmt.Errorf("unrecognized message type: %d", msg.Type)
	}

	log.Println("Message validation successful")
	return nil
}

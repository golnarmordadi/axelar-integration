// This code is implementing middleware for the Inter-Blockchain Communication (IBC) protocol in a
// Cosmos-SDK-based blockchain application. Middleware is a layer that intercepts or modifies
// the behavior of IBC communication while delegating core tasks to the underlying application
// module (app)

// How This code Works
// 1. Channel Lifecycle:
// Channels are created between two blockchains for communication.
// Middleware intercepts lifecycle events, logging actions and delegating to the underlying module.
// 2. Packet Reception:
// Middleware processes incoming packets.
// Extracts relevant data (Sender, Amount, Memo).
// If the Type in the memo is recognized (e.g., GeneralMessage), invokes custom logic in handler.
// 3. Custom Message Processing:
// Example: If the memo contains a GeneralMessageWithToken, the middleware:
// Parses the amount and token denomination.
// Invokes the HandleGeneralMessageWithToken method in handler.

package gmp_middleware

import (
	"encoding/json"
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
)

type IBCMiddleware struct {
	app     porttypes.IBCModule
	handler GeneralMessageHandler
}

// NewIBCMiddleware creates a new instance of IBCMiddleware.
func NewIBCMiddleware(app porttypes.IBCModule, handler GeneralMessageHandler) IBCMiddleware {
	log.Println("Initializing IBC Middleware")
	return IBCMiddleware{
		app:     app,
		handler: handler,
	}
}

// OnChanOpenInit handles the initialization of a new IBC channel.
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	log.Printf("OnChanOpenInit called with portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)
}

// OnChanOpenTry handles the handshake process of opening an IBC channel.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	log.Printf("OnChanOpenTry called with portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck acknowledges the opening of a channel.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	log.Printf("OnChanOpenAck called with portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm confirms the opening of a channel.
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	log.Printf("OnChanOpenConfirm called with portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit handles the initiation of a channel closure.
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	log.Printf("OnChanCloseInit called for portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm handles the confirmation of a channel closure.
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	log.Printf("OnChanCloseConfirm called for portID: %s, channelID: %s", portID, channelID)
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket processes an incoming packet.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	log.Printf("OnRecvPacket called with packet sequence: %d", packet.Sequence)

	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		log.Printf("Failed to process packet: %v", ack)
		return ack
	}

	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		log.Printf("Error unmarshalling packet data: %v", err)
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed to unmarshal ICS-20 transfer packet data: %w", err))
	}

	if data.Sender != AxelarGMPAcc {
		log.Printf("Unauthorized sender: %s", data.Sender)
		return ack
	}

	var msg Message
	if err := json.Unmarshal([]byte(data.GetMemo()), &msg); err != nil {
		log.Printf("Error unmarshalling memo: %v", err)
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("failed to unmarshal memo: %w", err))
	}

	var err error
	switch msg.Type {
	case TypeGeneralMessage:
		err = im.handler.HandleGeneralMessage(ctx, msg.SourceChain, msg.SourceAddress, data.Receiver, msg.Payload)
	case TypeGeneralMessageWithToken:
		amt, ok := sdk.NewIntFromString(data.Amount)
		if !ok {
			log.Printf("Invalid transfer amount: %s", data.Amount)
			return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "invalid transfer amount: %s", data.Amount))
		}

		denom := parseDenom(packet, data.Denom)
		err = im.handler.HandleGeneralMessageWithToken(ctx, msg.SourceChain, msg.SourceAddress, data.Receiver, msg.Payload, sdk.NewCoin(denom, amt))
	default:
		err = fmt.Errorf("unrecognized message type: %d", msg.Type)
		log.Printf("Error: %v", err)
	}

	if err != nil {
		log.Printf("Error processing message: %v", err)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	log.Println("Packet successfully processed")
	return ack
}

// OnAcknowledgementPacket handles packet acknowledgments.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	log.Printf("OnAcknowledgementPacket called for packet sequence: %d", packet.Sequence)
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket handles packet timeouts.
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	log.Printf("OnTimeoutPacket called for packet sequence: %d", packet.Sequence)
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

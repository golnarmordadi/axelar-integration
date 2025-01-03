package gmp_middleware

import (
	"encoding/json"
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/port/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
)

// IBCMiddleware wraps an IBC application with additional middleware logic.
type IBCMiddleware struct {
	app     porttypes.IBCModule
	handler GeneralMessageHandler
}

// NewIBCMiddleware creates a new IBCMiddleware instance.
func NewIBCMiddleware(app porttypes.IBCModule, handler GeneralMessageHandler) IBCMiddleware {
	log.Println("Initializing IBC Middleware...")
	return IBCMiddleware{
		app:     app,
		handler: handler,
	}
}

// OnChanOpenInit handles the initialization of a new channel.
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
	log.Printf("OnChanOpenInit called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)
}

// OnChanOpenTry handles the handshake phase of opening a channel.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	log.Printf("OnChanOpenTry called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck acknowledges the opening of a channel.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID, counterpartyChannelID, counterpartyVersion string,
) error {
	log.Printf("OnChanOpenAck called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm confirms the opening of a channel.
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID, channelID string,
) error {
	log.Printf("OnChanOpenConfirm called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit initiates the closure of a channel.
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID, channelID string,
) error {
	log.Printf("OnChanCloseInit called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm confirms the closure of a channel.
func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID, channelID string,
) error {
	log.Printf("OnChanCloseConfirm called: portID=%s, channelID=%s", portID, channelID)
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket handles incoming packets.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	log.Printf("OnRecvPacket called: packet sequence=%d", packet.Sequence)

	// Delegate to the underlying application.
	ack := im.app.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		log.Printf("Underlying application failed to process packet: %v", ack)
		return ack
	}

	var data transfertypes.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		log.Printf("Error unmarshalling packet data: %v", err)
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal ICS-20 transfer packet data: %w", err))
	}

	// Validate the sender address.
	if data.Sender != AxelarGMPAcc {
		log.Printf("Unauthorized sender: %s", data.Sender)
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unauthorized sender: %s", data.Sender))
	}

	var msg Message
	if err := json.Unmarshal([]byte(data.GetMemo()), &msg); err != nil {
		log.Printf("Error unmarshalling memo: %v", err)
		return channeltypes.NewErrorAcknowledgement(fmt.Errorf("cannot unmarshal memo: %w", err))
	}

	var err error
	switch msg.Type {
	case TypeGeneralMessage:
		log.Println("Processing GeneralMessage...")
		err = im.handler.HandleGeneralMessage(ctx, msg.SourceChain, msg.SourceAddress, data.Receiver, msg.Payload)
	case TypeGeneralMessageWithToken:
		log.Println("Processing GeneralMessageWithToken...")
		amt, ok := sdk.NewIntFromString(data.Amount)
		if !ok {
			log.Printf("Invalid amount: %s", data.Amount)
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

	log.Println("Packet processed successfully")
	return ack
}

// OnAcknowledgementPacket processes acknowledgments.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	log.Printf("OnAcknowledgementPacket called: packet sequence=%d", packet.Sequence)
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket handles packet timeouts.
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	log.Printf("OnTimeoutPacket called: packet sequence=%d", packet.Sequence)
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

# Requirements

1- **Be IBC-enabled:** Include IBC modules in the chain.
2- **Be x/ibc-transfer-enabled:** Chain should already have the x/ibc-transfer module enabled.
3- **Enabled ICS-20:** Implement the ICS-20 Token Transfer standard for transferring assets via IBC.
4- **IBC Relayers**
5- **Add Axelar-Specific IBC Configuration**
6- **Channel Setup**
7- **Connection Parameters**

# CrossChain Transfer

Axelar enables general messages passing between EVM and Cosmos chains.

In a nutshell, Axelar verifies messages originating from an EVM chain and forwards them as memo fields through ICS20 packets.

For chains with enabled Wasm modules, Axelar supports calling a CosmWasm contract from an EVM smart contract.

Two integration methods are currently supported:

1- **Native Network Integration:** Axelar verifies and forwards arbitrary payloads to the destination Cosmos chain. The receiving chain must implement an IBC middleware with a custom handler to decode and process the payload.

2- **EVM Networks Integration:** For chains with enabled Wasm modules, Axelar supports calling a CosmWasm contract from an EVM smart contract. The receiving chain must install the general-purpose IBC hook middleware. Message sender have the option to either encode the payload using Axelar's defined schema or pass the JSON CosmWasm contract call directly. Axelar verifies the payload and translates it into a Wasm execute message.

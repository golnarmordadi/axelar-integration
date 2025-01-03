# Step 1: Prerequisites

1- Set Up Cosmos Blockchain:

Ensure your blockchain supports IBC and the ICS-20 token transfer standard.
Your chain should already have the x/ibc-transfer module enabled.

2- Install Dependencies:

Install Axelar's Go SDK to interact with Axelar's APIs:
bash
Copy code
go get github.com/axelarnetwork/axelar-go-sdk

3- Set Up Axelar Testnet or Mainnet:

Use Axelar's public testnet for testing or mainnet for production:
Testnet RPC: https://rpc.testnet.axelar.dev
Mainnet RPC: https://rpc.axelar.dev

4- Relayer Configuration:

Install and configure the IBC relayer (e.g., hermes or relayer CLI) to manage packet forwarding between your chain and Axelar.

# Step 2: Setting Up the IBC Channel

```sh
hermes create channel --a-chain chain-id --a-port transfer --b-chain axelar-testnet --b-port transfer --order unordered --version ics20-1
```

```sh
hermes query channels --chain chain-id
```

# Step 3: Sending Tokens to Axelar

# Step 4: Verifying the Transfer

```sh
chain-cli query tx <tx-hash>
```
'use strict'

const {
    providers: { JsonRpcProvider },
    Contract,
    Wallet,
} = require('ethers')

const config = require('config');
const [ganache] = config.get('chains');

const Send = require('./artifacts/contracts/SendReceive.sol/SendReceive.json');
const IERC20 = require('./artifacts/@axelar-network/axelar-gmp-sdk-solidity/contracts/interfaces/IERC20.sol/IERC20.json');

// TODO: Create correct info
const tokenAddr = '';
const contract = '';

// TODO: Create correct arge
const destChain = '';
const destAddress = '';
const receiver = [''];
const symbol = '';
const amount = 1000000;

(async () => {
    const wallet = new Wallet(
        ganache.privateKey,
        new JsonRpcProvider(ganache.url),
    );
    
    const send = new Contract(contract, Send.abi, wallet);
    const usda = new Contract(tokenAddr, IERC20.abi, wallet);

    console.log(`wallet has ${(await usda.balanceOf(wallet.address)) / 1e6} ${symbol}`)
    console.log(`gateway is ${(await send.gateway())}`)

    const approveTx = await usda.approve(send.address, amount);
    await approveTx.wait();

    const sendTx = await send.multiSend(destChain, destAddress, receiver, symbol, amount);
    const tx = await sendTx.wait();
    
    console.log(`transaction hash is ${tx.transactionHash}`);
})();


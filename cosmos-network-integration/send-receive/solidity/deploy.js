'use strict'

const {
    providers: { JsonRpcProvider },
    Wallet,
    ContractFactory,
    constants: { AddressZero },
} = require('ethers')

const config = require('config');
const [ganache] = config.get('chains');

const Send = require('./artifacts/contracts/SendReceive.sol/SendReceive.json');


(async () => {
    const wallet = new Wallet(
        ganache.privateKey,
        new JsonRpcProvider(ganache.url),
    );
    const factory = ContractFactory.fromSolidity(Send, wallet);

    const contract = await factory.deploy(ganache.gateway, AddressZero)
    const tx = await contract.deployed();

    console.log(`send receive contract deployed on ${tx.address}`);

})();


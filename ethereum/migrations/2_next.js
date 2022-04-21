require("dotenv").config();

const ScaleCodec = artifacts.require("ScaleCodec");
const ETHApp = artifacts.require("ETHApp");
const ERC20App = artifacts.require("ERC20App");
const DOTApp = artifacts.require("DOTApp");
const TestToken = artifacts.require("TestToken");
const ValidatorRegistry = artifacts.require("ValidatorRegistry");
const MerkleProof = artifacts.require("MerkleProof");
const Bitfield = artifacts.require("Bitfield");

const channels = {
  basic: {
    inbound: {
      contract: artifacts.require("BasicInboundChannel"),
      instance: null
    },
    outbound: {
      contract: artifacts.require("BasicOutboundChannel"),
      instance: null,
    }
  },
  incentivized: {
    inbound: {
      contract: artifacts.require("IncentivizedInboundChannel"),
      instance: null
    },
    outbound: {
      contract: artifacts.require("IncentivizedOutboundChannel"),
      instance: null
    }
  },
}

const contracts = {
  lightclientbridge: {
    contract: artifacts.require("LightClientBridge"),
    instance: null
  }
}

module.exports = function (deployer, network, accounts) {
  deployer.then(async () => {

    // Account of governance contract
    // TODO: deploy the contract in this migration and use its address
    let administrator = accounts[0];

    // Fee for incentivized channel
    if (!("INCENTIVIZED_CHANNEL_FEE" in process.env)) {
      throw "Missing INCENTIVIZED_CHANNEL_FEE in environment config"
    }
    const fee = process.env.INCENTIVIZED_CHANNEL_FEE

    channels.basic.inbound.instance = await deployer.deploy(channels.basic.inbound.contract)
    channels.basic.outbound.instance = await deployer.deploy(channels.basic.outbound.contract)
    channels.incentivized.inbound.instance = await deployer.deploy(channels.incentivized.inbound.contract)
    channels.incentivized.outbound.instance = await deployer.deploy(channels.incentivized.outbound.contract)

    // Link libraries to applications
    await deployer.deploy(ScaleCodec);
    deployer.link(ScaleCodec, [ETHApp, ERC20App, DOTApp]);

    // Deploy applications
    const ethApp = await deployer.deploy(
      ETHApp,
      {
        inbound: channels.basic.inbound.instance.address,
        outbound: channels.basic.outbound.instance.address,
      },
      {
        inbound: channels.incentivized.inbound.instance.address,
        outbound: channels.incentivized.outbound.instance.address,
      },
    );

    const erc20App = await deployer.deploy(
      ERC20App,
      {
        inbound: channels.basic.inbound.instance.address,
        outbound: channels.basic.outbound.instance.address,
      },
      {
        inbound: channels.incentivized.inbound.instance.address,
        outbound: channels.incentivized.outbound.instance.address,
      },
    );

    const token = await deployer.deploy(TestToken, "Test Token", "TEST");

    // Deploy ERC1820 Registry for our E2E stack.
    if (network === 'e2e_test') {

      require('@openzeppelin/test-helpers/configure')({ web3 });
      const { singletons } = require('@openzeppelin/test-helpers');

      await singletons.ERC1820Registry(accounts[0]);
    }

    // only deploy this contract to non-development networks. The unit tests deploy this contract themselves.
    if (network === 'ropsten' || network === 'e2e_test') {
      const dotApp = await deployer.deploy(
        DOTApp,
        "Snowfork DOT",
        "SnowDOT",
        channels.incentivized.outbound.instance.address,
        {
          inbound: channels.basic.inbound.instance.address,
          outbound: channels.basic.outbound.instance.address,
        },
        {
          inbound: channels.incentivized.inbound.instance.address,
          outbound: channels.incentivized.outbound.instance.address,
        },
      );

      // Do post-construction initialization.
      await channels.incentivized.outbound.instance.initialize(
        administrator,
        dotApp.address,
        [dotApp.address, ethApp.address, erc20App.address]
      );
      await channels.incentivized.outbound.instance.setFee(
        fee,
        { from: administrator }
      );
    }

    await token.mint("10000", {
      from: accounts[0],
    });
    await token.mint("10000", {
      from: accounts[1],
    });

    // Link MerkleProof library to ValidatorRegistry
    await deployer.deploy(MerkleProof);
    deployer.link(MerkleProof, [ValidatorRegistry]);

    // TODO: Hardcoded for testing
    const root = "0xc1490f71b21f5700063d93546dbe860cc190e883734ee3f490b76de9e028db99";
    const numValidators = 2;
    const valRegistry = await deployer.deploy(ValidatorRegistry, root, numValidators);

    // Link Bitfield library to LightClientBridge
    await deployer.deploy(Bitfield);
    deployer.link(Bitfield, [contracts.lightclientbridge.contract]);

    contracts.lightclientbridge.instance = await deployer.deploy(contracts.lightclientbridge.contract, valRegistry.address)

  })
};

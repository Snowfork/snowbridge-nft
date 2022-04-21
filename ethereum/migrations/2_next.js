const ScaleCodec = artifacts.require("ScaleCodec");
const ETHApp = artifacts.require("ETHApp");
const ERC20App = artifacts.require("ERC20App");
const DOTAppDecimals10 = artifacts.require("DOTAppDecimals10");
const DOTAppDecimals12 = artifacts.require("DOTAppDecimals12");
const TestToken = artifacts.require("TestToken");

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

module.exports = function(deployer, network, accounts) {
  deployer.then(async () => {
    channels.basic.inbound.instance = await deployer.deploy(channels.basic.inbound.contract)
    channels.basic.outbound.instance = await deployer.deploy(channels.basic.outbound.contract)
    channels.incentivized.inbound.instance = await deployer.deploy(channels.incentivized.inbound.contract)
    channels.incentivized.outbound.instance = await deployer.deploy(channels.incentivized.outbound.contract)

    // Link libraries to applications
    await deployer.deploy(ScaleCodec);
    deployer.link(ScaleCodec, [ETHApp, ERC20App, DOTAppDecimals10, DOTAppDecimals12]);

    // Deploy applications
    await deployer.deploy(
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

    await deployer.deploy(
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

    await deployer.deploy(TestToken, 100000000, "Test Token", "TEST");

    // Deploy ERC1820 Registry for our E2E stack.
    if (network === 'e2e_test')  {

      require('@openzeppelin/test-helpers/configure')({ web3 });
      const { singletons } = require('@openzeppelin/test-helpers');

      await singletons.ERC1820Registry(accounts[0]);
    }

    // only deploy this contract to non-development networks. The unit tests deploy this contract themselves.
    if (network === 'ropsten' || network === 'e2e_test')  {
      await deployer.deploy(
        DOTAppDecimals12, // On Kusama and Rococo, KSM/ROC tokens have 12 decimal places
        "Snowfork DOT",
        "SnowDOT",
        {
          inbound: channels.basic.inbound.instance.address,
          outbound: channels.basic.outbound.instance.address,
        },
        {
          inbound: channels.incentivized.inbound.instance.address,
          outbound: channels.incentivized.outbound.instance.address,
        },
      );
    }

  })
};

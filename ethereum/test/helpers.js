
const deployAppWithMockChannels = async (deployer, channels, appContract, ...appContractArgs) => {
  const app = await appContract.new(
    ...appContractArgs,
    {
      inbound: channels[0],
      outbound: channels[1],
    },
    {
      inbound: channels[0],
      outbound: channels[1],
    },
    {
      from: deployer,
    }
  );

  return app;
}

const ChannelId = {
  Basic: 0,
  Incentivized: 1,
}

module.exports = {
  deployAppWithMockChannels,
  ChannelId
};

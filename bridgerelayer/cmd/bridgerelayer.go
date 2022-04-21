package cmd

import (
	"fmt"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/snowfork/polkadot-ethereum/bridgerelayer/cmd/chains/ethereum"
	// "github.com/snowfork/polkadot-ethereum/bridgerelayer/cmd/chains/substrate"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:          "bridgerelayer",
	Short:        "Streams transactions from Ethereum and Polkadot and relays tx information to the opposite chain",
	SilenceUsage: true,
}

//	initRelayerCmd
func initRelayerCmd() *cobra.Command {
	//nolint:lll
	initRelayerCmd := &cobra.Command{
		Use:   "init [polkadotRpcURL] [ethereumRpcUrl]",
		Short: "Validate credentials and initialize subscriptions to both chains",
		Args:  cobra.ExactArgs(2),
		// Example: "bridgerelayer init wss://rpc.polkadot.io wss://mainnet.infura.io/ws/v3/${INFURA_PROJECT_URL}",
		Example: "bridgerelayer init not_implemented ws://localhost:7545/",
		RunE:    RunInitRelayerCmd,
	}

	return initRelayerCmd
}

// RunInitRelayerCmd executes initRelayerCmd
func RunInitRelayerCmd(cmd *cobra.Command, args []string) error {
	// Validate and parse arguments
	if len(strings.Trim(args[0], "")) == 0 {
		return errors.Errorf("invalid [polkadot-rpc-url]: %s", args[0])
	}
	polkadotRPCURL := args[0]

	if len(strings.Trim(args[1], "")) == 0 {
		return errors.Errorf("invalid [ethereum-rpc-url]: %s", args[1])
	}
	ethereumRPCURL := args[1]

	// Initialize Ethereum chain
	ethStreamer := ethereum.NewStreamer(ethereumRPCURL)
	ethRouter := ethereum.NewRouter()
	ethChain := ethereum.NewEthChain(ethStreamer, ethRouter)

	// Initialize Substrate chain
	// subStreamer := substrate.NewStreamer(polkadotRPCURL)
	// subRouter := substrate.NewRouter()
	// subChain := substrate.NewSubChain(subStreamer, subRouter)
	_ = polkadotRPCURL

	// Start chains
	ethChain.Start()
	// subChain.Start()

	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bridgerelayer.yaml)")

	// Construct Root Command
	rootCmd.AddCommand(
		initRelayerCmd(),
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".bridgerelayer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".bridgerelayer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

package ethereum

import (
	log "github.com/sirupsen/logrus"

	"sync"

	"github.com/snowfork/polkadot-ethereum/bridgerelayer/types"
)

// EthChain streams the Ethereum blockchain and routes tx data packets
type EthChain struct {
	Streamer Streamer // The streamer of this chain
	Router   Router   // The router of this chain
}

// NewEthChain initializes a new instance of EthChain
func NewEthChain(streamer Streamer, router Router) EthChain {
	return EthChain{
		Streamer: streamer,
		Router:   router,
	}
}

func (ec EthChain) Start(wg *sync.WaitGroup) error {
	defer wg.Done()

	errors := make(chan error, 0)
	events := make(chan types.EventData, 0)

	go ec.Streamer.Start(events, errors)

	for {
		select {
		case err := <-errors:
			log.Error(err)
		case event := <-events:
			err := ec.Router.Route(event)
			if err != nil {
				log.Error(err)
			}
		}
	}
}

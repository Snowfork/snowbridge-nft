// Copyright 2020 Snowfork
// SPDX-License-Identifier: LGPL-3.0-only

package ethereum

import (
	"context"

	geth "github.com/ethereum/go-ethereum"
	gethCommon "github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/snowfork/polkadot-ethereum/bridgerelayer/chain"
)

// Listener streams the Ethereum blockchain for application events
type Listener struct {
	conn     *Connection
	apps     []Application
	messages chan<- chain.Message
	log      *logrus.Entry
}

func NewListener(conn *Connection, messages chan<- chain.Message, log *logrus.Entry) (*Listener, error) {
	apps := LoadApplications(viper.GetString("ethereum.registry-path"))

	return &Listener{
		conn:     conn,
		apps:     apps,
		messages: messages,
		log:      log,
	}, nil
}

func (li *Listener) Start(cxt context.Context, eg *errgroup.Group) error {
	eg.Go(func() error {
		return li.pollEvents(cxt)
	})

	return nil
}

func (li *Listener) pollEvents(ctx context.Context) error {
	li.log.Info("Polling started")

	events := make(chan gethTypes.Log)
	for _, app := range li.apps {
		query := makeQuery(app)

		_, err := li.conn.client.SubscribeFilterLogs(ctx, query, events)
		if err != nil {
			li.log.WithFields(logrus.Fields{
				"address": app.ID,
			}).Error("Failed to subscribe to application events")
			continue
		}

		li.log.WithFields(logrus.Fields{
			"address": app.ID,
		}).Info("Subscribed to application events")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-events:
			li.log.WithFields(logrus.Fields{
				"address":     event.Address.Hex(),
				"txHash":      event.TxHash.Hex(),
				"blockNumber": event.BlockNumber,
			}).Info("Witnessed transaction for application")

			msg, err := MakeMessageFromEvent(event, li.conn.kp)
			if err != nil {
				li.log.WithFields(logrus.Fields{
					"address":     event.Address.Hex(),
					"txHash":      event.TxHash.Hex(),
					"blockNumber": event.BlockNumber,
				}).Error("Failed to generate message from ethereum event")
			} else {
				li.messages <- *msg
			}
		}
	}
}

func makeQuery(app Application) geth.FilterQuery {
	address := gethCommon.HexToAddress(app.ID)
	signature := app.ABI.Events[EventName].ID.Hex()
	topic := gethCommon.HexToHash(signature)

	return geth.FilterQuery{
		Addresses: []gethCommon.Address{address},
		Topics:    [][]gethCommon.Hash{{topic}},
	}
}

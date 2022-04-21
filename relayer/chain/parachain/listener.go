// Copyright 2020 Snowfork
// SPDX-License-Identifier: LGPL-3.0-only

package parachain

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	rpcOffchain "github.com/snowfork/go-substrate-rpc-client/v2/rpc/offchain"
	"github.com/snowfork/go-substrate-rpc-client/v2/types"
	"golang.org/x/sync/errgroup"

	"github.com/snowfork/polkadot-ethereum/relayer/chain"
	chainTypes "github.com/snowfork/polkadot-ethereum/relayer/substrate"
)

type Listener struct {
	config   *Config
	conn     *Connection
	messages chan<- []chain.Message
	log      *logrus.Entry
}

func NewListener(config *Config, conn *Connection, messages chan<- []chain.Message, log *logrus.Entry) *Listener {
	return &Listener{
		config:   config,
		conn:     conn,
		messages: messages,
		log:      log,
	}
}

func (li *Listener) Start(ctx context.Context, eg *errgroup.Group) error {
	eg.Go(func() error {
		return li.pollBlocks(ctx)
	})

	return nil
}

func (li *Listener) onDone(ctx context.Context) error {
	li.log.Info("Shutting down listener...")
	close(li.messages)
	return ctx.Err()
}

func (li *Listener) pollBlocks(ctx context.Context) error {
	if li.messages == nil {
		li.log.Info("Not polling events since channel is nil")
		return nil
	}

	// Get current block
	block, err := li.conn.api.RPC.Chain.GetHeaderLatest()
	if err != nil {
		return err
	}
	currentBlock := uint32(block.Number)

	retryInterval := time.Duration(10) * time.Second
	for {
		select {
		case <-ctx.Done():
			return li.onDone(ctx)
		default:

			li.log.WithField("block", currentBlock).Debug("Processing block")

			// Get block hash
			finalizedHash, err := li.conn.api.RPC.Chain.GetFinalizedHead()
			if err != nil {
				li.log.WithError(err).Error("Failed to fetch finalized head")
				sleep(ctx, retryInterval)
				continue
			}

			// Get block header
			finalizedHeader, err := li.conn.api.RPC.Chain.GetHeader(finalizedHash)
			if err != nil {
				li.log.WithError(err).Error("Failed to fetch header for finalized head")
				sleep(ctx, retryInterval)
				continue
			}

			// Sleep if the block we want comes after the most recently finalized block
			if currentBlock > uint32(finalizedHeader.Number) {
				li.log.WithFields(logrus.Fields{
					"block":  currentBlock,
					"latest": finalizedHeader.Number,
				}).Trace("Block not yet finalized")
				sleep(ctx, retryInterval)
				continue
			}

			digestItem, err := getAuxiliaryDigestItem(finalizedHeader.Digest)
			if err != nil {
				return err
			}

			if digestItem != nil && digestItem.IsCommitment {
				li.log.WithFields(logrus.Fields{
					"block":          finalizedHeader.Number,
					"channelID":      digestItem.AsCommitment.ChannelID,
					"commitmentHash": digestItem.AsCommitment.Hash.Hex(),
				}).Debug("Found commitment hash in header digest")

				storageKey, err := MakeStorageKey(digestItem.AsCommitment.ChannelID, digestItem.AsCommitment.Hash)
				if err != nil {
					return err
				}

				data, err := li.conn.api.RPC.Offchain.LocalStorageGet(rpcOffchain.Persistent, storageKey)
				if err != nil {
					li.log.WithError(err).Error("Failed to read commitment from offchain storage")
					sleep(ctx, retryInterval)
					continue
				}

				if data != nil {
					li.log.WithFields(logrus.Fields{
						"block":               finalizedHeader.Number,
						"commitmentSizeBytes": len(*data),
					}).Debug("Retrieved commitment from offchain storage")
				} else {
					li.log.WithError(err).Error("Commitment not found in offchain storage")
					continue
				}

				var messages []chainTypes.CommitmentMessage

				err = types.DecodeFromBytes(*data, &messages)
				if err != nil {
					li.log.WithError(err).Error("Faild to decode commitment messages")
				}

				message := chain.SubstrateOutboundMessage{
					ChannelID:      digestItem.AsCommitment.ChannelID,
					CommitmentHash: digestItem.AsCommitment.Hash,
					Commitment:     messages,
				}

				li.messages <- []chain.Message{message}
			}

			currentBlock++
		}
	}
}

func sleep(ctx context.Context, delay time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

func getAuxiliaryDigestItem(digest types.Digest) (*chainTypes.AuxiliaryDigestItem, error) {
	for _, digestItem := range digest {
		if digestItem.IsOther {
			var auxDigestItem chainTypes.AuxiliaryDigestItem
			err := types.DecodeFromBytes(digestItem.AsOther, &auxDigestItem)
			if err != nil {
				return nil, err
			}
			return &auxDigestItem, nil
		}
	}
	return nil, nil
}

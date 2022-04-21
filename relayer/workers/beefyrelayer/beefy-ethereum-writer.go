package beefyrelayer

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"

	"github.com/snowfork/polkadot-ethereum/relayer/chain/ethereum"
	"github.com/snowfork/polkadot-ethereum/relayer/contracts/lightclientbridge"
	"github.com/snowfork/polkadot-ethereum/relayer/workers/beefyrelayer/store"
)

type BeefyEthereumWriter struct {
	ethereumConfig    *ethereum.Config
	ethereumConn      *ethereum.Connection
	beefyDB           *store.Database
	lightClientBridge *lightclientbridge.Contract
	databaseMessages  chan<- store.DatabaseCmd
	beefyMessages     <-chan store.BeefyRelayInfo
	log               *logrus.Entry
}

func NewBeefyEthereumWriter(ethereumConfig *ethereum.Config, ethereumConn *ethereum.Connection, beefyDB *store.Database,
	databaseMessages chan<- store.DatabaseCmd, beefyMessages <-chan store.BeefyRelayInfo,
	log *logrus.Entry) *BeefyEthereumWriter {
	return &BeefyEthereumWriter{
		ethereumConfig:   ethereumConfig,
		ethereumConn:     ethereumConn,
		beefyDB:          beefyDB,
		databaseMessages: databaseMessages,
		beefyMessages:    beefyMessages,
		log:              log,
	}
}

func (wr *BeefyEthereumWriter) Start(ctx context.Context, eg *errgroup.Group) error {

	lightClientBridgeContract, err := lightclientbridge.NewContract(common.HexToAddress(wr.ethereumConfig.LightClientBridge), wr.ethereumConn.GetClient())
	if err != nil {
		return err
	}
	wr.lightClientBridge = lightClientBridgeContract

	eg.Go(func() error {
		return wr.writeMessagesLoop(ctx)
	})

	return nil
}

func (wr *BeefyEthereumWriter) onDone(ctx context.Context) error {
	wr.log.Info("Shutting down writer...")
	// Avoid deadlock if a listener is still trying to send to a channel
	for range wr.beefyMessages {
		wr.log.Debug("Discarded BEEFY message")
	}
	return ctx.Err()
}

func (wr *BeefyEthereumWriter) writeMessagesLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return wr.onDone(ctx)
		case msg := <-wr.beefyMessages:
			switch msg.Status {
			case store.CommitmentWitnessed:
				err := wr.WriteNewSignatureCommitment(ctx, msg, 0) // TODO: pick val addr
				if err != nil {
					wr.log.WithError(err).Error("Error submitting message to ethereum")
				}
			case store.ReadyToComplete:
				err := wr.WriteCompleteSignatureCommitment(ctx, msg)
				if err != nil {
					wr.log.WithError(err).Error("Error submitting message to ethereum")
				}
			}
		}
	}
}

func (wr *BeefyEthereumWriter) signerFn(_ common.Address, tx *types.Transaction) (*types.Transaction, error) {
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, wr.ethereumConn.GetKP().PrivateKey())
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

func (wr *BeefyEthereumWriter) WriteNewSignatureCommitment(ctx context.Context, info store.BeefyRelayInfo, valIndex int) error {
	beefyJustification, err := info.ToBeefyJustification()
	if err != nil {
		return fmt.Errorf("Error converting BeefyRelayInfo to BeefyJustification: %s", err.Error())
	}

	msg, err := beefyJustification.BuildNewSignatureCommitmentMessage(valIndex)
	if err != nil {
		return err
	}

	contract := wr.lightClientBridge
	if contract == nil {
		return fmt.Errorf("Unknown contract")
	}

	options := bind.TransactOpts{
		From:     wr.ethereumConn.GetKP().CommonAddress(),
		Signer:   wr.signerFn,
		Context:  ctx,
		GasLimit: 5000000,
	}

	tx, err := contract.NewSignatureCommitment(&options, msg.CommitmentHash,
		msg.ValidatorClaimsBitfield, msg.ValidatorSignatureCommitment,
		msg.ValidatorPosition, msg.ValidatorPublicKey, msg.ValidatorPublicKeyMerkleProof)
	if err != nil {
		wr.log.WithError(err).Error("Failed to submit transaction")
		return err
	}

	wr.log.WithFields(logrus.Fields{
		"txHash": tx.Hash().Hex(),
	}).Info("New Signature Commitment transaction submitted")

	wr.log.Info("2: Updating item in Database with status 'InitialVerificationTxSent'")
	instructions := map[string]interface{}{
		"status":                       store.InitialVerificationTxSent,
		"initial_verification_tx_hash": tx.Hash(),
	}
	cmd := store.NewDatabaseCmd(&info, store.Update, instructions)
	wr.databaseMessages <- cmd

	return nil
}

// WriteCompleteSignatureCommitment sends a CompleteSignatureCommitment tx to the LightClientBridge contract
func (wr *BeefyEthereumWriter) WriteCompleteSignatureCommitment(ctx context.Context, info store.BeefyRelayInfo) error {
	beefyJustification, err := info.ToBeefyJustification()
	if err != nil {
		return fmt.Errorf("Error converting BeefyRelayInfo to BeefyJustification: %s", err.Error())
	}

	msg, err := beefyJustification.BuildCompleteSignatureCommitmentMessage()
	if err != nil {
		return err
	}

	contract := wr.lightClientBridge
	if contract == nil {
		return fmt.Errorf("Unknown contract")
	}

	options := bind.TransactOpts{
		From:     wr.ethereumConn.GetKP().CommonAddress(),
		Signer:   wr.signerFn,
		Context:  ctx,
		GasLimit: 500000,
	}

	tx, err := contract.CompleteSignatureCommitment(&options, msg.ID, msg.CommitmentHash, msg.Commitment, msg.Signatures,
		msg.ValidatorPositions, msg.ValidatorPublicKeys, msg.ValidatorPublicKeyMerkleProofs)

	if err != nil {
		wr.log.WithError(err).Error("Failed to submit transaction")
		return err
	}

	wr.log.WithFields(logrus.Fields{
		"txHash": tx.Hash().Hex(),
	}).Info("Complete Signature Commitment transaction submitted")

	// Update item's status in database
	wr.log.Info("5: Updating item status from 'ReadyToComplete' to 'CompleteVerificationTxSent'")
	instructions := map[string]interface{}{
		"status":                        store.CompleteVerificationTxSent,
		"complete_verification_tx_hash": tx.Hash(),
	}
	updateCmd := store.NewDatabaseCmd(&info, store.Update, instructions)
	wr.databaseMessages <- updateCmd

	return nil
}

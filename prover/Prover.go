package prover

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
)

// Prover can verify transactions via a SPV Proof or Threshold Vote
type Prover interface {
	ProofVerifier

	// VerifyTransaction passes a transaction to ProofVerifier of ThresholdVerifier
	VerifyTransaction(verificationData VerificationData) (bool, error)
}

// VerificationData is the base interface for different verification strategies
type VerificationData interface{}

// ProofVerifier verifies transactions via a SPV Proof
type ProofVerifier interface {
	// BuildProof builds a SPV proof
	BuildProof(block types.Block, txHash string) SpvProof

	// VerifyProof verifies a SPV proof by proving commitment to its root (not to block hash)
	VerifyProof(merklePath string, tx string, parentNodes string, header string, blockHash string) bool
}

// LightClientProof supports cryptographic verification
type LightClientProof struct {
	verificationData VerificationData
	proof            SpvProof // An SPV proof
}

// SpvProof contains information used to verify that a transaction is included in a specific block
type SpvProof struct {
	blockHash   common.Hash  // Block hash
	blockHeader types.Header // Block header
	parentNodes []string
	txIndex     int           // Transaction's position in a block
	txReceipt   types.Receipt // Raw transaction receipt
}
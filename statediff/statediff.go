package statediff

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type StateDiff struct {
	BlockNumber     int64                                     `json:"blockNumber"      gencodec:"required"`
	BlockHash       common.Hash                               `json:"blockHash"        gencodec:"required"`
	CreatedAccounts map[common.Address]AccountDiffEventual    `json:"createdAccounts"  gencodec:"required"`
	DeletedAccounts map[common.Address]AccountDiffEventual    `json:"deletedAccounts"  gencodec:"required"`
	UpdatedAccounts map[common.Address]AccountDiffIncremental `json:"updatedAccounts"  gencodec:"required"`

	encoded []byte
	err     error
}

func (self *StateDiff) ensureEncoded() {
	if self.encoded == nil && self.err == nil {
		self.encoded, self.err = json.Marshal(self)
	}
}

// Implement Encoder interface for StateDiff
func (sd *StateDiff) Length() int {
	sd.ensureEncoded()
	return len(sd.encoded)
}

// Implement Encoder interface for StateDiff
func (sd *StateDiff) Encode() ([]byte, error) {
	sd.ensureEncoded()
	return sd.encoded, sd.err
}

type AccountDiffEventual struct {
	Nonce        diffUint64            `json:"nonce"         gencodec:"required"`
	Balance      diffBigInt            `json:"balance"       gencodec:"required"`
	Code         string                `json:"code"          gencodec:"required"`
	CodeHash     string                `json:"codeHash"      gencodec:"required"`
	ContractRoot diffString            `json:"contractRoot"  gencodec:"required"`
	Storage      map[string]diffString `json:"storage"       gencodec:"required"`
}

type AccountDiffIncremental struct {
	Nonce        diffUint64            `json:"nonce"         gencodec:"required"`
	Balance      diffBigInt            `json:"balance"       gencodec:"required"`
	CodeHash     string                `json:"codeHash"      gencodec:"required"`
	ContractRoot diffString            `json:"contractRoot"  gencodec:"required"`
	Storage      map[string]diffString `json:"storage"       gencodec:"required"`
}

type diffString struct {
	NewValue *string `json:"newValue"  gencodec:"optional"`
	OldValue *string `json:"oldValue"  gencodec:"optional"`
}
type diffUint64 struct {
	NewValue *uint64 `json:"newValue"  gencodec:"optional"`
	OldValue *uint64 `json:"oldValue"  gencodec:"optional"`
}
type diffBigInt struct {
	NewValue *big.Int `json:"newValue"  gencodec:"optional"`
	OldValue *big.Int `json:"oldValue"  gencodec:"optional"`
}

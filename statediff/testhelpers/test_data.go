package testhelpers

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/statediff/builder"
	"math/big"
	"math/rand"
)

var (
	BlockNumber     = rand.Int63()
	BlockHash       = "0xfa40fbe2d98d98b3363a778d52f2bcd29d6790b9b3f3cab2b167fd12d3550f73"
	CodeHash        = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	OldNonceValue   = rand.Uint64()
	NewNonceValue   = OldNonceValue + 1
	OldBalanceValue = rand.Int63()
	NewBalanceValue = OldBalanceValue - 1
	ContractRoot    = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
	StoragePath     = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	oldStorage      = "0x0"
	newStorage      = "0x03"
	storage         = map[string]builder.DiffString{StoragePath: {
		NewValue: &newStorage,
		OldValue: &oldStorage,
	}}
	address             = common.HexToAddress("0xaE9BEa628c4Ce503DcFD7E305CaB4e29E7476592")
	CreatedAccountDiffs = map[common.Address]builder.AccountDiffEventual{address: {
		Nonce: builder.DiffUint64{
			NewValue: &NewNonceValue,
			OldValue: &OldNonceValue,
		},
		Balance: builder.DiffBigInt{
			NewValue: big.NewInt(NewBalanceValue),
			OldValue: big.NewInt(OldBalanceValue),
		},
		ContractRoot: builder.DiffString{
			NewValue: &ContractRoot,
			OldValue: &ContractRoot,
		},
		Code:     []byte("created account code"),
		CodeHash: CodeHash,
		Storage:  storage,
	}}

	UpdatedAccountDiffs = map[common.Address]builder.AccountDiffIncremental{address: {
		Nonce: builder.DiffUint64{
			NewValue: &NewNonceValue,
			OldValue: &OldNonceValue,
		},
		Balance: builder.DiffBigInt{
			NewValue: big.NewInt(NewBalanceValue),
			OldValue: big.NewInt(OldBalanceValue),
		},
		CodeHash: CodeHash,
		ContractRoot: builder.DiffString{
			NewValue: &ContractRoot,
			OldValue: &ContractRoot,
		},
		Storage: storage,
	}}

	DeletedAccountDiffs = map[common.Address]builder.AccountDiffEventual{address: {
		Nonce: builder.DiffUint64{
			NewValue: &NewNonceValue,
			OldValue: &OldNonceValue,
		},
		Balance: builder.DiffBigInt{
			NewValue: big.NewInt(NewBalanceValue),
			OldValue: big.NewInt(OldBalanceValue),
		},
		ContractRoot: builder.DiffString{
			NewValue: &ContractRoot,
			OldValue: &ContractRoot,
		},
		Code:     []byte("deleted account code"),
		CodeHash: CodeHash,
		Storage:  storage,
	}}

	TestStateDiff = builder.StateDiff{
		BlockNumber:     BlockNumber,
		BlockHash:       common.HexToHash(BlockHash),
		CreatedAccounts: CreatedAccountDiffs,
		DeletedAccounts: DeletedAccountDiffs,
		UpdatedAccounts: UpdatedAccountDiffs,
	}
)

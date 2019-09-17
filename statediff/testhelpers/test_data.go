// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package testhelpers

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/statediff"
)

// AddressToLeafKey hashes an returns an address
func AddressToLeafKey(address common.Address) common.Hash {
	return common.BytesToHash(crypto.Keccak256(address[:]))
}

// Test variables
var (
	BlockNumber     = big.NewInt(rand.Int63())
	BlockHash       = "0xfa40fbe2d98d98b3363a778d52f2bcd29d6790b9b3f3cab2b167fd12d3550f73"
	CodeHash        = common.Hex2Bytes("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	NewNonceValue   = rand.Uint64()
	NewBalanceValue = rand.Int63()
	ContractRoot    = common.HexToHash("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
	StoragePath     = common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes()
	StorageKey      = common.HexToHash("0000000000000000000000000000000000000000000000000000000000000001").Bytes()
	StorageValue    = common.Hex2Bytes("0x03")
	storage         = []statediff.StorageDiff{{
		Key:   StorageKey,
		Value: StorageValue,
		Path:  StoragePath,
		Proof: [][]byte{},
	}}
	emptyStorage           = make([]statediff.StorageDiff, 0)
	address                = common.HexToAddress("0xaE9BEa628c4Ce503DcFD7E305CaB4e29E7476592")
	ContractLeafKey        = AddressToLeafKey(address)
	anotherAddress         = common.HexToAddress("0xaE9BEa628c4Ce503DcFD7E305CaB4e29E7476593")
	AnotherContractLeafKey = AddressToLeafKey(anotherAddress)
	testAccount            = state.Account{
		Nonce:    NewNonceValue,
		Balance:  big.NewInt(NewBalanceValue),
		Root:     ContractRoot,
		CodeHash: CodeHash,
	}
	valueBytes, _       = rlp.EncodeToBytes(testAccount)
	CreatedAccountDiffs = []statediff.AccountDiff{
		{
			Key:     ContractLeafKey.Bytes(),
			Value:   valueBytes,
			Storage: storage,
		},
		{
			Key:     AnotherContractLeafKey.Bytes(),
			Value:   valueBytes,
			Storage: emptyStorage,
		},
	}

	UpdatedAccountDiffs = []statediff.AccountDiff{{
		Key:     ContractLeafKey.Bytes(),
		Value:   valueBytes,
		Storage: storage,
	}}

	DeletedAccountDiffs = []statediff.AccountDiff{{
		Key:     ContractLeafKey.Bytes(),
		Value:   valueBytes,
		Storage: storage,
	}}

	TestStateDiff = statediff.StateDiff{
		BlockNumber:     BlockNumber,
		BlockHash:       common.HexToHash(BlockHash),
		CreatedAccounts: CreatedAccountDiffs,
		DeletedAccounts: DeletedAccountDiffs,
		UpdatedAccounts: UpdatedAccountDiffs,
	}
	Testdb = rawdb.NewMemoryDatabase()

	TestBankKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	TestBankAddress = crypto.PubkeyToAddress(TestBankKey.PublicKey) //0x71562b71999873DB5b286dF957af199Ec94617F7
	BankLeafKey     = AddressToLeafKey(TestBankAddress)
	TestBankFunds   = big.NewInt(100000000)
	Genesis         = core.GenesisBlockForTesting(Testdb, TestBankAddress, TestBankFunds)

	Account1Key, _  = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	Account2Key, _  = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	Account1Addr    = crypto.PubkeyToAddress(Account1Key.PublicKey) //0x703c4b2bD70c169f5717101CaeE543299Fc946C7
	Account2Addr    = crypto.PubkeyToAddress(Account2Key.PublicKey) //0x0D3ab14BBaD3D99F4203bd7a11aCB94882050E7e
	Account1LeafKey = AddressToLeafKey(Account1Addr)
	Account2LeafKey = AddressToLeafKey(Account2Addr)
	ContractCode    = common.Hex2Bytes("608060405234801561001057600080fd5b50600160036000600a811061002157fe5b01819055506102a9806100356000396000f3fe608060405234801561001057600080fd5b50600436106100575760003560e01c806320b682e51461005c57806338b48bcd1461008a5780639126f719146100cc578063d0eb18d6146100d6578063ebda8d4014610120575b600080fd5b61006461013e565b604051808265ffffffffffff1665ffffffffffff16815260200191505060405180910390f35b6100b6600480360360208110156100a057600080fd5b8101908080359060200190929190505050610156565b6040518082815260200191505060405180910390f35b6100d4610174565b005b6100de610248565b604051808273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200191505060405180910390f35b61012861026e565b6040518082815260200191505060405180910390f35b600260149054906101000a900465ffffffffffff1681565b60016020528060005260406000206000915090508060000154905081565b7f7465737420646174610000000000000000000000000000000000000000000000600081905550736c3abb37949f1e009baf32fc91b69513769974d5600260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555060028060146101000a81548165ffffffffffff021916908365ffffffffffff1602179055506003806000600a811061022457fe5b01819055506004600160006001815260200190815260200160002060000181905550565b600260009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000548156fea265627a7a723058203ab132ed8cfb8dcb5fae041ad021f8fd2d57056144d4819fa123f1cdbcbc068b64736f6c634300050a0032")
	ContractAddr    common.Address
)

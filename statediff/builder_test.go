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

package statediff_test

import (
	"bytes"
	"math/big"
	"sort"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
)

var (
	contractLeafKey                common.Hash
	emptyAccountDiffEventualMap    = make([]statediff.AccountDiff, 0)
	emptyAccountDiffIncrementalMap = make([]statediff.AccountDiff, 0)
	block0, block1, block2, block3 *types.Block
	builder                        statediff.Builder
	miningReward                   = int64(2000000000000000000)
	burnAddress                    = common.HexToAddress("0x0")
	burnLeafKey                    = testhelpers.AddressToLeafKey(burnAddress)
	block0Hash                     = common.HexToHash("0xd1721cfd0b29c36fd7a68f25c128e86413fb666a6e1d68e89b875bd299262661")
	block1Hash                     = common.HexToHash("0xbbe88de60ba33a3f18c0caa37d827bfb70252e19e40a07cd34041696c35ecb1a")
	block2Hash                     = common.HexToHash("0x0538f299356e9c4dd86d59f647fd409ea848be3cedff648a9b1b933660341eec")
	block3Hash                     = common.HexToHash("0x9a375813e362ff0a25102b30530f34ec4cceaa36a8a5d0b88396394457a57eab")
	balanceChange10000             = int64(10000)
	balanceChange1000              = int64(1000)
	block1BankBalance              = int64(99990000)
	block1Account1Balance          = int64(10000)
	block2Account2Balance          = int64(1000)
	nonce0                         = uint64(0)
	nonce1                         = uint64(1)
	nonce2                         = uint64(2)
	nonce3                         = uint64(3)
	originalContractRoot           = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
	contractStorageRoot            = "0x85b0d1a99eb49a28eb6e29db9788e518aabb2b2d0bbd475868f8ad34581770c0"
	newContractRoot                = "0x7cc40eefd6e1f91d01a7d3015c7687da166d9ec454b6cf277840b8755309452d"

	//slot 0: bytes32Data
	storageSlotZero  = common.HexToHash("0")
	storageSlotZeroKey             = crypto.Keccak256Hash(storageSlotZero[:])                                                                                   //TODO: rename to storageSlotZeroKey
	updatedBytes32DataStorageValue = []byte{160, 116, 101, 115, 116, 32, 100, 97, 116, 97, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} //this is the []byte representation of "test data"

	//slot 1: mapping(uint => TestStruct)
	// calculate key for TestStruct.var1
	mappingKey      = "1"
	mappingKeyBytes = []byte{1}
	indexInContract = "0000000000000000000000000000000000000000000000000000000000000001"
	leftPaddedBytes = common.LeftPadBytes(mappingKeyBytes, 32)
	hexKey          = common.Bytes2Hex(leftPaddedBytes)
	keyBytes        =common.FromHex(hexKey + indexInContract)
	encoded         = crypto.Keccak256Hash(keyBytes)        // cc69885fda6bcc1a4ace058b4a62bf5e179ea78fd58a1ccd71c22cc9b688792f
	keccakOfKey     = crypto.Keccak256Hash(encoded.Bytes()) // the statediff service is currently emitting the key as a keccakhas

	testStructVar1Value = common.Hex2Bytes("04")

	//slot 2: addressData and uint48Data (since the address and uint48 data are both <32 bytes, they are packed into one storage slot
	storageSlotTwo   = common.HexToHash("2")
	storageSlotTwoKey                 = crypto.Keccak256Hash(storageSlotTwo[:]).Bytes()
	storageOneSlotRlpEncodeValue = []byte{149, 2, 108, 58, 187, 55, 148, 159, 30, 0, 155, 175, 50, 252, 145, 182, 149, 19, 118, 153, 116, 213}

	// slot 3: uintArrayData
	storageSlotThree = common.HexToHash("3")
	storageSlotThreeKey            = crypto.Keccak256Hash(storageSlotThree[:]).Bytes()
	originalUintArrayDataStorageValue = common.Hex2Bytes("01")
	updatedUintArrayDataStorageValue  = common.Hex2Bytes("03")



	account1Block1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(balanceChange10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	burnAccountBlock1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	bankAccountBlock1, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(testhelpers.TestBankFunds.Int64() - balanceChange10000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account2Block2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	contractAccountBlock2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0x16121d4252af839f48ea17ab4bf8e8a3c9130e59582427fbf7af8879ae54aa49").Bytes(),
		Root:     common.HexToHash(contractStorageRoot),
	})
	bankAccountBlock2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce2,
		Balance:  big.NewInt(block1BankBalance - balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account1Block2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce2,
		Balance:  big.NewInt(block1Account1Balance - balanceChange1000 + balanceChange1000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	burnAccountBlock2, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(miningReward + miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	account2Block3, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce0,
		Balance:  big.NewInt(block2Account2Balance + miningReward),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
	contractAccountBlock3, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce1,
		Balance:  big.NewInt(0),
		CodeHash: common.HexToHash("0x16121d4252af839f48ea17ab4bf8e8a3c9130e59582427fbf7af8879ae54aa49").Bytes(),
		Root:     common.HexToHash(newContractRoot),
	})
	bankAccountBlock3, _ = rlp.EncodeToBytes(state.Account{
		Nonce:    nonce3,
		Balance:  big.NewInt(99989000),
		CodeHash: common.HexToHash("0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470").Bytes(),
		Root:     common.HexToHash(originalContractRoot),
	})
)

type arguments struct {
	oldStateRoot common.Hash
	newStateRoot common.Hash
	blockNumber  *big.Int
	blockHash    common.Hash
}

func TestBuilder(t *testing.T) {
	_, blockMap, chain := testhelpers.MakeChain(3, testhelpers.Genesis)
	contractLeafKey = testhelpers.AddressToLeafKey(testhelpers.ContractAddr)
	defer chain.Stop()
	block0 = blockMap[block0Hash]
	block1 = blockMap[block1Hash]
	block2 = blockMap[block2Hash]
	block3 = blockMap[block3Hash]
	config := statediff.Config{
		PathsAndProofs:    true,
		IntermediateNodes: false,
	}
	builder = statediff.NewBuilder(testhelpers.Testdb, chain, config)

	var tests = []struct {
		name              string
		startingArguments arguments
		expected          *statediff.StateDiff
	}{
		{
			"testEmptyDiff",
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block0.Root(),
				blockNumber:  block0.Number(),
				blockHash:    block0Hash,
			},
			&statediff.StateDiff{
				BlockNumber:     block0.Number(),
				BlockHash:       block0Hash,
				CreatedAccounts: emptyAccountDiffEventualMap,
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: emptyAccountDiffIncrementalMap,
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block1.Root(),
				blockNumber:  block1.Number(),
				blockHash:    block1Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   burnLeafKey.Bytes(),
						Value: burnAccountBlock1,
						Proof: [][]byte{{248, 113, 160, 87, 118, 82, 182, 37, 183, 123, 219, 91, 247, 123, 196, 63, 49, 37, 202, 215, 70, 77, 103, 157, 21, 117, 86, 82, 119, 211, 97, 27, 128, 83, 231, 128, 128, 128, 128, 160, 254, 136, 159, 16, 229, 219, 143, 44, 43, 243, 85, 146, 129, 82, 161, 127, 110, 59, 185, 154, 146, 65, 172, 109, 132, 199, 126, 98, 100, 80, 156, 121, 128, 128, 128, 128, 128, 128, 128, 128, 160, 17, 219, 12, 218, 52, 168, 150, 218, 190, 182, 131, 155, 176, 106, 56, 244, 149, 20, 207, 164, 134, 67, 89, 132, 235, 1, 59, 125, 249, 238, 133, 197, 128, 128},
							{248, 113, 160, 51, 128, 199, 183, 174, 129, 165, 142, 185, 141, 156, 120, 222, 74, 31, 215, 253, 149, 53, 252, 149, 62, 210, 190, 96, 45, 170, 164, 23, 103, 49, 42, 184, 78, 248, 76, 128, 136, 27, 193, 109, 103, 78, 200, 0, 0, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{5, 3, 8, 0, 12, 7, 11, 7, 10, 14, 8, 1, 10, 5, 8, 14, 11, 9, 8, 13, 9, 12, 7, 8, 13, 14, 4, 10, 1, 15, 13, 7, 15, 13, 9, 5, 3, 5, 15, 12, 9, 5, 3, 14, 13, 2, 11, 14, 6, 0, 2, 13, 10, 10, 10, 4, 1, 7, 6, 7, 3, 1, 2, 10, 16},
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1Block1,
						Proof: [][]byte{{248, 113, 160, 87, 118, 82, 182, 37, 183, 123, 219, 91, 247, 123, 196, 63, 49, 37, 202, 215, 70, 77, 103, 157, 21, 117, 86, 82, 119, 211, 97, 27, 128, 83, 231, 128, 128, 128, 128, 160, 254, 136, 159, 16, 229, 219, 143, 44, 43, 243, 85, 146, 129, 82, 161, 127, 110, 59, 185, 154, 146, 65, 172, 109, 132, 199, 126, 98, 100, 80, 156, 121, 128, 128, 128, 128, 128, 128, 128, 128, 160, 17, 219, 12, 218, 52, 168, 150, 218, 190, 182, 131, 155, 176, 106, 56, 244, 149, 20, 207, 164, 134, 67, 89, 132, 235, 1, 59, 125, 249, 238, 133, 197, 128, 128},
							{248, 107, 160, 57, 38, 219, 105, 170, 206, 213, 24, 233, 185, 240, 244, 52, 164, 115, 231, 23, 65, 9, 201, 67, 84, 139, 184, 242, 59, 228, 28, 167, 109, 154, 210, 184, 72, 248, 70, 128, 130, 39, 16, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{14, 9, 2, 6, 13, 11, 6, 9, 10, 10, 12, 14, 13, 5, 1, 8, 14, 9, 11, 9, 15, 0, 15, 4, 3, 4, 10, 4, 7, 3, 14, 7, 1, 7, 4, 1, 0, 9, 12, 9, 4, 3, 5, 4, 8, 11, 11, 8, 15, 2, 3, 11, 14, 4, 1, 12, 10, 7, 6, 13, 9, 10, 13, 2, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccountBlock1,
						Proof: [][]byte{{248, 113, 160, 87, 118, 82, 182, 37, 183, 123, 219, 91, 247, 123, 196, 63, 49, 37, 202, 215, 70, 77, 103, 157, 21, 117, 86, 82, 119, 211, 97, 27, 128, 83, 231, 128, 128, 128, 128, 160, 254, 136, 159, 16, 229, 219, 143, 44, 43, 243, 85, 146, 129, 82, 161, 127, 110, 59, 185, 154, 146, 65, 172, 109, 132, 199, 126, 98, 100, 80, 156, 121, 128, 128, 128, 128, 128, 128, 128, 128, 160, 17, 219, 12, 218, 52, 168, 150, 218, 190, 182, 131, 155, 176, 106, 56, 244, 149, 20, 207, 164, 134, 67, 89, 132, 235, 1, 59, 125, 249, 238, 133, 197, 128, 128},
							{248, 109, 160, 48, 191, 73, 244, 64, 161, 205, 5, 39, 228, 208, 110, 39, 101, 101, 76, 15, 86, 69, 34, 87, 81, 109, 121, 58, 155, 141, 96, 77, 207, 223, 42, 184, 74, 248, 72, 1, 132, 5, 245, 185, 240, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{0, 0, 11, 15, 4, 9, 15, 4, 4, 0, 10, 1, 12, 13, 0, 5, 2, 7, 14, 4, 13, 0, 6, 14, 2, 7, 6, 5, 6, 5, 4, 12, 0, 15, 5, 6, 4, 5, 2, 2, 5, 7, 5, 1, 6, 13, 7, 9, 3, 10, 9, 11, 8, 13, 6, 0, 4, 13, 12, 15, 13, 15, 2, 10, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock2",
			//1000 transferred from testBankAddress to account1Addr
			//1000 transferred from account1Addr to account2Addr
			arguments{
				oldStateRoot: block1.Root(),
				newStateRoot: block2.Root(),
				blockNumber:  block2.Number(),
				blockHash:    block2Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccountBlock2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 105, 160, 49, 20, 101, 138, 116, 217, 204, 159, 122, 207, 44, 92, 214, 150, 195, 73, 77, 124, 52, 77, 120, 191, 236, 58, 221, 13, 145, 236, 78, 141, 28, 69, 184, 70, 248, 68, 1, 128, 160, 133, 176, 209, 169, 158, 180, 154, 40, 235, 110, 41, 219, 151, 136, 229, 24, 170, 187, 43, 45, 11, 189, 71, 88, 104, 248, 173, 52, 88, 23, 112, 192, 160, 22, 18, 29, 66, 82, 175, 131, 159, 72, 234, 23, 171, 75, 248, 232, 163, 201, 19, 14, 89, 88, 36, 39, 251, 247, 175, 136, 121, 174, 84, 170, 73}},
						Path: []byte{6, 1, 1, 4, 6, 5, 8, 10, 7, 4, 13, 9, 12, 12, 9, 15, 7, 10, 12, 15, 2, 12, 5, 12, 13, 6, 9, 6, 12, 3, 4, 9, 4, 13, 7, 12, 3, 4, 4, 13, 7, 8, 11, 15, 14, 12, 3, 10, 13, 13, 0, 13, 9, 1, 14, 12, 4, 14, 8, 13, 1, 12, 4, 5, 16},
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   storageSlotThreeKey,
								Value: originalUintArrayDataStorageValue,
								Proof: [][]byte{{227, 161, 32, 194, 87, 90, 14, 158, 89, 60, 0, 249, 89, 248, 201, 47, 18, 219, 40, 105, 195, 57, 90, 59, 5, 2, 208, 94, 37, 22, 68, 111, 113, 248, 91, 1}},
								Path: []byte{12, 2, 5, 7, 5, 10, 0, 14, 9, 14, 5, 9, 3, 12, 0, 0, 15, 9, 5, 9, 15, 8, 12, 9, 2, 15, 1, 2, 13, 11, 2, 8, 6, 9, 12, 3, 3, 9, 5, 10, 3, 11, 0, 5, 0, 2, 13, 0, 5, 14, 2, 5, 1, 6, 4, 4, 6, 15, 7, 1, 15, 8, 5, 11, 16},
							},
						},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account2LeafKey.Bytes(),
						Value: account2Block2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 107, 160, 57, 87, 243, 226, 240, 74, 7, 100, 195, 160, 73, 27, 23, 95, 105, 146, 109, 166, 30, 251, 204, 143, 97, 250, 20, 85, 253, 45, 43, 76, 221, 69, 184, 72, 248, 70, 128, 130, 3, 232, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{12, 9, 5, 7, 15, 3, 14, 2, 15, 0, 4, 10, 0, 7, 6, 4, 12, 3, 10, 0, 4, 9, 1, 11, 1, 7, 5, 15, 6, 9, 9, 2, 6, 13, 10, 6, 1, 14, 15, 11, 12, 12, 8, 15, 6, 1, 15, 10, 1, 4, 5, 5, 15, 13, 2, 13, 2, 11, 4, 12, 13, 13, 4, 5, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccountBlock2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 109, 160, 48, 191, 73, 244, 64, 161, 205, 5, 39, 228, 208, 110, 39, 101, 101, 76, 15, 86, 69, 34, 87, 81, 109, 121, 58, 155, 141, 96, 77, 207, 223, 42, 184, 74, 248, 72, 2, 132, 5, 245, 182, 8, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{0, 0, 11, 15, 4, 9, 15, 4, 4, 0, 10, 1, 12, 13, 0, 5, 2, 7, 14, 4, 13, 0, 6, 14, 2, 7, 6, 5, 6, 5, 4, 12, 0, 15, 5, 6, 4, 5, 2, 2, 5, 7, 5, 1, 6, 13, 7, 9, 3, 10, 9, 11, 8, 13, 6, 0, 4, 13, 12, 15, 13, 15, 2, 10, 16},
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   burnLeafKey.Bytes(),
						Value: burnAccountBlock2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 113, 160, 51, 128, 199, 183, 174, 129, 165, 142, 185, 141, 156, 120, 222, 74, 31, 215, 253, 149, 53, 252, 149, 62, 210, 190, 96, 45, 170, 164, 23, 103, 49, 42, 184, 78, 248, 76, 128, 136, 55, 130, 218, 206, 157, 144, 0, 0, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{5, 3, 8, 0, 12, 7, 11, 7, 10, 14, 8, 1, 10, 5, 8, 14, 11, 9, 8, 13, 9, 12, 7, 8, 13, 14, 4, 10, 1, 15, 13, 7, 15, 13, 9, 5, 3, 5, 15, 12, 9, 5, 3, 14, 13, 2, 11, 14, 6, 0, 2, 13, 10, 10, 10, 4, 1, 7, 6, 7, 3, 1, 2, 10, 16},
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1Block2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 107, 160, 57, 38, 219, 105, 170, 206, 213, 24, 233, 185, 240, 244, 52, 164, 115, 231, 23, 65, 9, 201, 67, 84, 139, 184, 242, 59, 228, 28, 167, 109, 154, 210, 184, 72, 248, 70, 2, 130, 39, 16, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{14, 9, 2, 6, 13, 11, 6, 9, 10, 10, 12, 14, 13, 5, 1, 8, 14, 9, 11, 9, 15, 0, 15, 4, 3, 4, 10, 4, 7, 3, 14, 7, 1, 7, 4, 1, 0, 9, 12, 9, 4, 3, 5, 4, 8, 11, 11, 8, 15, 2, 3, 11, 14, 4, 1, 12, 10, 7, 6, 13, 9, 10, 13, 2, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			arguments{
				oldStateRoot: block2.Root(),
				newStateRoot: block3.Root(),
				blockNumber:  block3.Number(),
				blockHash:    block3.Hash(),
			},
			&statediff.StateDiff{
				BlockNumber:     block3.Number(),
				BlockHash:       block3.Hash(),
				CreatedAccounts: []statediff.AccountDiff{},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.BankLeafKey.Bytes(),
						Value: bankAccountBlock3,
						Proof: [][]byte{{248, 177, 160, 101, 223, 138, 81, 34, 40, 229, 170, 198, 188, 136, 99, 7, 55, 33, 112, 160, 111, 181, 131, 167, 201, 131, 24, 201, 211, 177, 30, 159, 229, 246, 6, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 199, 15, 230, 126, 225, 0, 151, 63, 140, 75, 33, 113, 23, 175, 121, 225, 167, 67, 227, 117, 123, 240, 139, 143, 187, 185, 205, 38, 62, 164, 227, 175, 128, 128, 128, 128, 128, 160, 4, 228, 121, 222, 255, 218, 60, 247, 15, 0, 34, 198, 28, 229, 180, 129, 109, 157, 68, 181, 248, 229, 200, 123, 29, 81, 145, 114, 90, 209, 205, 210, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 109, 160, 48, 191, 73, 244, 64, 161, 205, 5, 39, 228, 208, 110, 39, 101, 101, 76, 15, 86, 69, 34, 87, 81, 109, 121, 58, 155, 141, 96, 77, 207, 223, 42, 184, 74, 248, 72, 3, 132, 5, 245, 182, 8, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{0, 0, 11, 15, 4, 9, 15, 4, 4, 0, 10, 1, 12, 13, 0, 5, 2, 7, 14, 4, 13, 0, 6, 14, 2, 7, 6, 5, 6, 5, 4, 12, 0, 15, 5, 6, 4, 5, 2, 2, 5, 7, 5, 1, 6, 13, 7, 9, 3, 10, 9, 11, 8, 13, 6, 0, 4, 13, 12, 15, 13, 15, 2, 10, 16},
						Storage: []statediff.StorageDiff{},
					},
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccountBlock3,
						Proof: [][]byte{{248, 177, 160, 101, 223, 138, 81, 34, 40, 229, 170, 198, 188, 136, 99, 7, 55, 33, 112, 160, 111, 181, 131, 167, 201, 131, 24, 201, 211, 177, 30, 159, 229, 246, 6, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 199, 15, 230, 126, 225, 0, 151, 63, 140, 75, 33, 113, 23, 175, 121, 225, 167, 67, 227, 117, 123, 240, 139, 143, 187, 185, 205, 38, 62, 164, 227, 175, 128, 128, 128, 128, 128, 160, 4, 228, 121, 222, 255, 218, 60, 247, 15, 0, 34, 198, 28, 229, 180, 129, 109, 157, 68, 181, 248, 229, 200, 123, 29, 81, 145, 114, 90, 209, 205, 210, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 105, 160, 49, 20, 101, 138, 116, 217, 204, 159, 122, 207, 44, 92, 214, 150, 195, 73, 77, 124, 52, 77, 120, 191, 236, 58, 221, 13, 145, 236, 78, 141, 28, 69, 184, 70, 248, 68, 1, 128, 160, 124, 196, 14, 239, 214, 225, 249, 29, 1, 167, 211, 1, 92, 118, 135, 218, 22, 109, 158, 196, 84, 182, 207, 39, 120, 64, 184, 117, 83, 9, 69, 45, 160, 22, 18, 29, 66, 82, 175, 131, 159, 72, 234, 23, 171, 75, 248, 232, 163, 201, 19, 14, 89, 88, 36, 39, 251, 247, 175, 136, 121, 174, 84, 170, 73}},
						Path: []byte{6, 1, 1, 4, 6, 5, 8, 10, 7, 4, 13, 9, 12, 12, 9, 15, 7, 10, 12, 15, 2, 12, 5, 12, 13, 6, 9, 6, 12, 3, 4, 9, 4, 13, 7, 12, 3, 4, 4, 13, 7, 8, 11, 15, 14, 12, 3, 10, 13, 13, 0, 13, 9, 1, 14, 12, 4, 14, 8, 13, 1, 12, 4, 5, 16},
						Storage: []statediff.StorageDiff{
							{ // slot 0: storage diff for bytes32Data
								Leaf:  true,
								Key:   storageSlotZeroKey[:],
								Value: updatedBytes32DataStorageValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{248, 67, 160, 57, 13, 236, 217, 84, 139, 98, 168, 214, 3, 69, 169, 136, 56, 111, 200, 75, 166, 188, 149, 72, 64, 8, 246, 54, 47, 147, 22, 14, 243, 229, 99, 161, 160, 116, 101, 115, 116, 32, 100, 97, 116, 97, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
								Path: []byte{2, 9, 0, 13, 14, 12, 13, 9, 5, 4, 8, 11, 6, 2, 10, 8, 13, 6, 0, 3, 4, 5, 10, 9, 8, 8, 3, 8, 6, 15, 12, 8, 4, 11, 10, 6, 11, 12, 9, 5, 4, 8, 4, 0, 0, 8, 15, 6, 3, 6, 2, 15, 9, 3, 1, 6, 0, 14, 15, 3, 14, 5, 6, 3, 16},
							},
							{ // slot 2: storage diff for address + uint48
								Leaf:  true,
								Key:   storageSlotTwoKey[:],
								Value: storageOneSlotRlpEncodeValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{248, 56, 160, 48, 87, 135, 250, 18, 168, 35, 224, 242, 183, 99, 28, 196, 27, 59, 168, 130, 139, 51, 33, 202, 129, 17, 17, 250, 117, 205, 58, 163, 187, 90, 206, 150, 149, 2, 108, 58, 187, 55, 148, 159, 30, 0, 155, 175, 50, 252, 145, 182, 149, 19, 118, 153, 116, 213}},
								Path: []byte{4, 0, 5, 7, 8, 7, 15, 10, 1, 2, 10, 8, 2, 3, 14, 0, 15, 2, 11, 7, 6, 3, 1, 12, 12, 4, 1, 11, 3, 11, 10, 8, 8, 2, 8, 11, 3, 3, 2, 1, 12, 10, 8, 1, 1, 1, 1, 1, 15, 10, 7, 5, 12, 13, 3, 10, 10, 3, 11, 11, 5, 10, 12, 14, 16},
							},
							{ //slot 1: storage diff for var 1 of TestStruct
								Leaf:  true,
								Key:   keccakOfKey[:],
								Value: testStructVar1Value,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{226, 160, 54, 179, 39, 64, 173, 128, 65, 188, 195, 185, 9, 199, 45, 126, 26, 254, 96, 9, 78, 197, 94, 60, 222, 50, 155, 75, 58, 40, 80, 29, 130, 108, 4}},
								Path: []byte{6, 6, 11, 3, 2, 7, 4, 0, 10, 13, 8, 0, 4, 1, 11, 12, 12, 3, 11, 9, 0, 9, 12, 7, 2, 13, 7, 14, 1, 10, 15, 14, 6, 0, 0, 9, 4, 14, 12, 5, 5, 14, 3, 12, 13, 14, 3, 2, 9, 11, 4, 11, 3, 10, 2, 8, 5, 0, 1, 13, 8, 2, 6, 12, 16},
							},
							{ // slot 3: storage diff for uintArrayData
								Leaf:  true,
								Key:   storageSlotThreeKey[:],
								Value: updatedUintArrayDataStorageValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{226, 160, 50, 87, 90, 14, 158, 89, 60, 0, 249, 89, 248, 201, 47, 18, 219, 40, 105, 195, 57, 90, 59, 5, 2, 208, 94, 37, 22, 68, 111, 113, 248, 91, 3}},
								Path: []byte{12, 2, 5, 7, 5, 10, 0, 14, 9, 14, 5, 9, 3, 12, 0, 0, 15, 9, 5, 9, 15, 8, 12, 9, 2, 15, 1, 2, 13, 11, 2, 8, 6, 9, 12, 3, 3, 9, 5, 10, 3, 11, 0, 5, 0, 2, 13, 0, 5, 14, 2, 5, 1, 6, 4, 4, 6, 15, 7, 1, 15, 8, 5, 11, 16},
							},
						},
					},
					{
						Leaf:  true,
						Key:   testhelpers.Account2LeafKey.Bytes(),
						Value: account2Block3,
						Proof: [][]byte{{248, 177, 160, 101, 223, 138, 81, 34, 40, 229, 170, 198, 188, 136, 99, 7, 55, 33, 112, 160, 111, 181, 131, 167, 201, 131, 24, 201, 211, 177, 30, 159, 229, 246, 6, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 199, 15, 230, 126, 225, 0, 151, 63, 140, 75, 33, 113, 23, 175, 121, 225, 167, 67, 227, 117, 123, 240, 139, 143, 187, 185, 205, 38, 62, 164, 227, 175, 128, 128, 128, 128, 128, 160, 4, 228, 121, 222, 255, 218, 60, 247, 15, 0, 34, 198, 28, 229, 180, 129, 109, 157, 68, 181, 248, 229, 200, 123, 29, 81, 145, 114, 90, 209, 205, 210, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 113, 160, 57, 87, 243, 226, 240, 74, 7, 100, 195, 160, 73, 27, 23, 95, 105, 146, 109, 166, 30, 251, 204, 143, 97, 250, 20, 85, 253, 45, 43, 76, 221, 69, 184, 78, 248, 76, 128, 136, 27, 193, 109, 103, 78, 200, 3, 232, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{12, 9, 5, 7, 15, 3, 14, 2, 15, 0, 4, 10, 0, 7, 6, 4, 12, 3, 10, 0, 4, 9, 1, 11, 1, 7, 5, 15, 6, 9, 9, 2, 6, 13, 10, 6, 1, 14, 15, 11, 12, 12, 8, 15, 6, 1, 15, 10, 1, 4, 5, 5, 15, 13, 2, 13, 2, 11, 4, 12, 13, 13, 4, 5, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		arguments := test.startingArguments
		diff, err := builder.BuildStateDiff(arguments.oldStateRoot, arguments.newStateRoot, arguments.blockNumber, arguments.blockHash)
		if err != nil {
			t.Error(err)
		}
		receivedStateDiffRlp, err := rlp.EncodeToBytes(diff)
		if err != nil {
			t.Error(err)
		}
		expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
		if err != nil {
			t.Error(err)
		}

		sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
		sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
		if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
			t.Logf("Test failed: %s", test.name)
			t.Errorf("actual state diff rlp: %+v\nexpected state diff rlp: %+v", receivedStateDiffRlp, expectedStateDiffRlp)
		}
	}
}

func TestBuilderWithWatchedAddressList(t *testing.T) {
	_, blockMap, chain := testhelpers.MakeChain(3, testhelpers.Genesis)
	contractLeafKey = testhelpers.AddressToLeafKey(testhelpers.ContractAddr)
	defer chain.Stop()
	block0 = blockMap[block0Hash]
	block1 = blockMap[block1Hash]
	block2 = blockMap[block2Hash]
	block3 = blockMap[block3Hash]
	config := statediff.Config{
		PathsAndProofs:    true,
		IntermediateNodes: false,
		WatchedAddresses:  []string{testhelpers.Account1Addr.Hex(), testhelpers.ContractAddr.Hex()},
	}
	builder = statediff.NewBuilder(testhelpers.Testdb, chain, config)

	var tests = []struct {
		name              string
		startingArguments arguments
		expected          *statediff.StateDiff
	}{
		{
			"testEmptyDiff",
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block0.Root(),
				blockNumber:  block0.Number(),
				blockHash:    block0Hash,
			},
			&statediff.StateDiff{
				BlockNumber:     block0.Number(),
				BlockHash:       block0Hash,
				CreatedAccounts: emptyAccountDiffEventualMap,
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: emptyAccountDiffIncrementalMap,
			},
		},
		{
			"testBlock1",
			//10000 transferred from testBankAddress to account1Addr
			arguments{
				oldStateRoot: block0.Root(),
				newStateRoot: block1.Root(),
				blockNumber:  block1.Number(),
				blockHash:    block1Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block1.Number(),
				BlockHash:   block1.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1Block1,
						Proof: [][]byte{{248, 113, 160, 87, 118, 82, 182, 37, 183, 123, 219, 91, 247, 123, 196, 63, 49, 37, 202, 215, 70, 77, 103, 157, 21, 117, 86, 82, 119, 211, 97, 27, 128, 83, 231, 128, 128, 128, 128, 160, 254, 136, 159, 16, 229, 219, 143, 44, 43, 243, 85, 146, 129, 82, 161, 127, 110, 59, 185, 154, 146, 65, 172, 109, 132, 199, 126, 98, 100, 80, 156, 121, 128, 128, 128, 128, 128, 128, 128, 128, 160, 17, 219, 12, 218, 52, 168, 150, 218, 190, 182, 131, 155, 176, 106, 56, 244, 149, 20, 207, 164, 134, 67, 89, 132, 235, 1, 59, 125, 249, 238, 133, 197, 128, 128},
							{248, 107, 160, 57, 38, 219, 105, 170, 206, 213, 24, 233, 185, 240, 244, 52, 164, 115, 231, 23, 65, 9, 201, 67, 84, 139, 184, 242, 59, 228, 28, 167, 109, 154, 210, 184, 72, 248, 70, 128, 130, 39, 16, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{14, 9, 2, 6, 13, 11, 6, 9, 10, 10, 12, 14, 13, 5, 1, 8, 14, 9, 11, 9, 15, 0, 15, 4, 3, 4, 10, 4, 7, 3, 14, 7, 1, 7, 4, 1, 0, 9, 12, 9, 4, 3, 5, 4, 8, 11, 11, 8, 15, 2, 3, 11, 14, 4, 1, 12, 10, 7, 6, 13, 9, 10, 13, 2, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{},
			},
		},
		{
			"testBlock2",
			//1000 transferred from testBankAddress to account1Addr
			//1000 transferred from account1Addr to account2Addr
			arguments{
				oldStateRoot: block1.Root(),
				newStateRoot: block2.Root(),
				blockNumber:  block2.Number(),
				blockHash:    block2Hash,
			},
			&statediff.StateDiff{
				BlockNumber: block2.Number(),
				BlockHash:   block2.Hash(),
				CreatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccountBlock2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 105, 160, 49, 20, 101, 138, 116, 217, 204, 159, 122, 207, 44, 92, 214, 150, 195, 73, 77, 124, 52, 77, 120, 191, 236, 58, 221, 13, 145, 236, 78, 141, 28, 69, 184, 70, 248, 68, 1, 128, 160, 133, 176, 209, 169, 158, 180, 154, 40, 235, 110, 41, 219, 151, 136, 229, 24, 170, 187, 43, 45, 11, 189, 71, 88, 104, 248, 173, 52, 88, 23, 112, 192, 160, 22, 18, 29, 66, 82, 175, 131, 159, 72, 234, 23, 171, 75, 248, 232, 163, 201, 19, 14, 89, 88, 36, 39, 251, 247, 175, 136, 121, 174, 84, 170, 73}},
						Path: []byte{6, 1, 1, 4, 6, 5, 8, 10, 7, 4, 13, 9, 12, 12, 9, 15, 7, 10, 12, 15, 2, 12, 5, 12, 13, 6, 9, 6, 12, 3, 4, 9, 4, 13, 7, 12, 3, 4, 4, 13, 7, 8, 11, 15, 14, 12, 3, 10, 13, 13, 0, 13, 9, 1, 14, 12, 4, 14, 8, 13, 1, 12, 4, 5, 16},
						Storage: []statediff.StorageDiff{
							{
								Leaf:  true,
								Key:   storageSlotThreeKey,
								Value: originalUintArrayDataStorageValue,
								Proof: [][]byte{{227, 161, 32, 194, 87, 90, 14, 158, 89, 60, 0, 249, 89, 248, 201, 47, 18, 219, 40, 105, 195, 57, 90, 59, 5, 2, 208, 94, 37, 22, 68, 111, 113, 248, 91, 1}},
								Path: []byte{12, 2, 5, 7, 5, 10, 0, 14, 9, 14, 5, 9, 3, 12, 0, 0, 15, 9, 5, 9, 15, 8, 12, 9, 2, 15, 1, 2, 13, 11, 2, 8, 6, 9, 12, 3, 3, 9, 5, 10, 3, 11, 0, 5, 0, 2, 13, 0, 5, 14, 2, 5, 1, 6, 4, 4, 6, 15, 7, 1, 15, 8, 5, 11, 16},
							},
						},
					},
				},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   testhelpers.Account1LeafKey.Bytes(),
						Value: account1Block2,
						Proof: [][]byte{{248, 177, 160, 177, 155, 238, 178, 242, 47, 83, 2, 49, 141, 155, 92, 149, 175, 245, 120, 233, 177, 101, 67, 46, 200, 23, 250, 41, 74, 135, 94, 61, 133, 51, 162, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 9, 107, 57, 187, 240, 243, 7, 160, 197, 170, 9, 243, 186, 60, 237, 49, 238, 93, 24, 81, 209, 59, 28, 186, 138, 100, 237, 220, 203, 160, 71, 148, 128, 128, 128, 128, 128, 160, 10, 173, 165, 125, 110, 240, 77, 112, 149, 100, 135, 237, 25, 228, 116, 7, 195, 9, 210, 166, 208, 148, 101, 23, 244, 238, 84, 84, 211, 249, 138, 137, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 107, 160, 57, 38, 219, 105, 170, 206, 213, 24, 233, 185, 240, 244, 52, 164, 115, 231, 23, 65, 9, 201, 67, 84, 139, 184, 242, 59, 228, 28, 167, 109, 154, 210, 184, 72, 248, 70, 2, 130, 39, 16, 160, 86, 232, 31, 23, 27, 204, 85, 166, 255, 131, 69, 230, 146, 192, 248, 110, 91, 72, 224, 27, 153, 108, 173, 192, 1, 98, 47, 181, 227, 99, 180, 33, 160, 197, 210, 70, 1, 134, 247, 35, 60, 146, 126, 125, 178, 220, 199, 3, 192, 229, 0, 182, 83, 202, 130, 39, 59, 123, 250, 216, 4, 93, 133, 164, 112}},
						Path:    []byte{14, 9, 2, 6, 13, 11, 6, 9, 10, 10, 12, 14, 13, 5, 1, 8, 14, 9, 11, 9, 15, 0, 15, 4, 3, 4, 10, 4, 7, 3, 14, 7, 1, 7, 4, 1, 0, 9, 12, 9, 4, 3, 5, 4, 8, 11, 11, 8, 15, 2, 3, 11, 14, 4, 1, 12, 10, 7, 6, 13, 9, 10, 13, 2, 16},
						Storage: []statediff.StorageDiff{},
					},
				},
			},
		},
		{
			"testBlock3",
			//the contract's storage is changed
			//and the block is mined by account 2
			arguments{
				oldStateRoot: block2.Root(),
				newStateRoot: block3.Root(),
				blockNumber:  block3.Number(),
				blockHash:    block3.Hash(),
			},
			&statediff.StateDiff{
				BlockNumber:     block3.Number(),
				BlockHash:       block3.Hash(),
				CreatedAccounts: []statediff.AccountDiff{},
				DeletedAccounts: emptyAccountDiffEventualMap,
				UpdatedAccounts: []statediff.AccountDiff{
					{
						Leaf:  true,
						Key:   contractLeafKey.Bytes(),
						Value: contractAccountBlock3,
						Proof: [][]byte{{248, 177, 160, 101, 223, 138, 81, 34, 40, 229, 170, 198, 188, 136, 99, 7, 55, 33, 112, 160, 111, 181, 131, 167, 201, 131, 24, 201, 211, 177, 30, 159, 229, 246, 6, 128, 128, 128, 128, 160, 179, 86, 53, 29, 96, 188, 152, 148, 207, 31, 29, 108, 182, 140, 129, 95, 1, 49, 213, 15, 29, 168, 60, 64, 35, 160, 158, 200, 85, 207, 255, 145, 160, 199, 15, 230, 126, 225, 0, 151, 63, 140, 75, 33, 113, 23, 175, 121, 225, 167, 67, 227, 117, 123, 240, 139, 143, 187, 185, 205, 38, 62, 164, 227, 175, 128, 128, 128, 128, 128, 160, 4, 228, 121, 222, 255, 218, 60, 247, 15, 0, 34, 198, 28, 229, 180, 129, 109, 157, 68, 181, 248, 229, 200, 123, 29, 81, 145, 114, 90, 209, 205, 210, 128, 160, 255, 115, 147, 190, 57, 135, 174, 188, 86, 51, 227, 70, 22, 253, 237, 49, 24, 19, 149, 199, 142, 195, 186, 244, 70, 51, 138, 0, 146, 148, 117, 60, 128, 128},
							{248, 105, 160, 49, 20, 101, 138, 116, 217, 204, 159, 122, 207, 44, 92, 214, 150, 195, 73, 77, 124, 52, 77, 120, 191, 236, 58, 221, 13, 145, 236, 78, 141, 28, 69, 184, 70, 248, 68, 1, 128, 160, 124, 196, 14, 239, 214, 225, 249, 29, 1, 167, 211, 1, 92, 118, 135, 218, 22, 109, 158, 196, 84, 182, 207, 39, 120, 64, 184, 117, 83, 9, 69, 45, 160, 22, 18, 29, 66, 82, 175, 131, 159, 72, 234, 23, 171, 75, 248, 232, 163, 201, 19, 14, 89, 88, 36, 39, 251, 247, 175, 136, 121, 174, 84, 170, 73}},
						Path: []byte{6, 1, 1, 4, 6, 5, 8, 10, 7, 4, 13, 9, 12, 12, 9, 15, 7, 10, 12, 15, 2, 12, 5, 12, 13, 6, 9, 6, 12, 3, 4, 9, 4, 13, 7, 12, 3, 4, 4, 13, 7, 8, 11, 15, 14, 12, 3, 10, 13, 13, 0, 13, 9, 1, 14, 12, 4, 14, 8, 13, 1, 12, 4, 5, 16},
						Storage: []statediff.StorageDiff{
							{ //storage diff for bytes32Data
								Leaf:  true,
								Key:   storageSlotZeroKey[:],
								Value: updatedBytes32DataStorageValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{248, 67, 160, 57, 13, 236, 217, 84, 139, 98, 168, 214, 3, 69, 169, 136, 56, 111, 200, 75, 166, 188, 149, 72, 64, 8, 246, 54, 47, 147, 22, 14, 243, 229, 99, 161, 160, 116, 101, 115, 116, 32, 100, 97, 116, 97, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
								Path: []byte{2, 9, 0, 13, 14, 12, 13, 9, 5, 4, 8, 11, 6, 2, 10, 8, 13, 6, 0, 3, 4, 5, 10, 9, 8, 8, 3, 8, 6, 15, 12, 8, 4, 11, 10, 6, 11, 12, 9, 5, 4, 8, 4, 0, 0, 8, 15, 6, 3, 6, 2, 15, 9, 3, 1, 6, 0, 14, 15, 3, 14, 5, 6, 3, 16},
							},
							{ // storage diff for address + uint48
								Leaf:  true,
								Key:   storageSlotTwoKey[:],
								Value: storageOneSlotRlpEncodeValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{248, 56, 160, 48, 87, 135, 250, 18, 168, 35, 224, 242, 183, 99, 28, 196, 27, 59, 168, 130, 139, 51, 33, 202, 129, 17, 17, 250, 117, 205, 58, 163, 187, 90, 206, 150, 149, 2, 108, 58, 187, 55, 148, 159, 30, 0, 155, 175, 50, 252, 145, 182, 149, 19, 118, 153, 116, 213}},
								Path: []byte{4, 0, 5, 7, 8, 7, 15, 10, 1, 2, 10, 8, 2, 3, 14, 0, 15, 2, 11, 7, 6, 3, 1, 12, 12, 4, 1, 11, 3, 11, 10, 8, 8, 2, 8, 11, 3, 3, 2, 1, 12, 10, 8, 1, 1, 1, 1, 1, 15, 10, 7, 5, 12, 13, 3, 10, 10, 3, 11, 11, 5, 10, 12, 14, 16},
							},
							{ //storage diff for var 1 of TestStruct
								Leaf:  true,
								Key:   keccakOfKey[:],
								Value: testStructVar1Value,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{226, 160, 54, 179, 39, 64, 173, 128, 65, 188, 195, 185, 9, 199, 45, 126, 26, 254, 96, 9, 78, 197, 94, 60, 222, 50, 155, 75, 58, 40, 80, 29, 130, 108, 4}},
								Path: []byte{6, 6, 11, 3, 2, 7, 4, 0, 10, 13, 8, 0, 4, 1, 11, 12, 12, 3, 11, 9, 0, 9, 12, 7, 2, 13, 7, 14, 1, 10, 15, 14, 6, 0, 0, 9, 4, 14, 12, 5, 5, 14, 3, 12, 13, 14, 3, 2, 9, 11, 4, 11, 3, 10, 2, 8, 5, 0, 1, 13, 8, 2, 6, 12, 16},
							},
							{ // storage diff for uintArrayData
								Leaf:  true,
								Key:   storageSlotThreeKey[:],
								Value: updatedUintArrayDataStorageValue,
								Proof: [][]byte{{248, 145, 128, 128, 160, 131, 156, 157, 229, 241, 229, 169, 135, 165, 29, 173, 181, 227, 247, 106, 24, 93, 54, 96, 54, 130, 34, 118, 15, 65, 136, 243, 57, 132, 179, 24, 15, 128, 160, 236, 14, 243, 12, 248, 114, 138, 99, 83, 220, 38, 86, 88, 72, 28, 50, 9, 125, 187, 191, 243, 60, 11, 73, 65, 86, 219, 100, 143, 31, 48, 21, 128, 160, 141, 43, 107, 231, 173, 34, 95, 14, 11, 236, 18, 80, 34, 182, 150, 149, 217, 19, 98, 95, 37, 77, 139, 251, 118, 174, 170, 130, 227, 57, 90, 212, 128, 128, 128, 128, 128, 160, 185, 43, 188, 252, 172, 173, 59, 131, 59, 77, 42, 73, 147, 6, 154, 243, 101, 184, 174, 31, 185, 74, 190, 92, 211, 248, 157, 151, 238, 145, 20, 98, 128, 128, 128, 128},
									{226, 160, 50, 87, 90, 14, 158, 89, 60, 0, 249, 89, 248, 201, 47, 18, 219, 40, 105, 195, 57, 90, 59, 5, 2, 208, 94, 37, 22, 68, 111, 113, 248, 91, 3}},
								Path: []byte{12, 2, 5, 7, 5, 10, 0, 14, 9, 14, 5, 9, 3, 12, 0, 0, 15, 9, 5, 9, 15, 8, 12, 9, 2, 15, 1, 2, 13, 11, 2, 8, 6, 9, 12, 3, 3, 9, 5, 10, 3, 11, 0, 5, 0, 2, 13, 0, 5, 14, 2, 5, 1, 6, 4, 4, 6, 15, 7, 1, 15, 8, 5, 11, 16},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		arguments := test.startingArguments
		diff, err := builder.BuildStateDiff(arguments.oldStateRoot, arguments.newStateRoot, arguments.blockNumber, arguments.blockHash)
		if err != nil {
			t.Error(err)
		}
		receivedStateDiffRlp, err := rlp.EncodeToBytes(diff)
		if err != nil {
			t.Error(err)
		}
		expectedStateDiffRlp, err := rlp.EncodeToBytes(test.expected)
		if err != nil {
			t.Error(err)
		}

		sort.Slice(receivedStateDiffRlp, func(i, j int) bool { return receivedStateDiffRlp[i] < receivedStateDiffRlp[j] })
		sort.Slice(expectedStateDiffRlp, func(i, j int) bool { return expectedStateDiffRlp[i] < expectedStateDiffRlp[j] })
		if !bytes.Equal(receivedStateDiffRlp, expectedStateDiffRlp) {
			t.Logf("Test failed: %s", test.name)
			t.Errorf("actual state diff rlp: %+v\nexpected state diff rlp: %+v", receivedStateDiffRlp, expectedStateDiffRlp)
		}
	}
}

/*
pragma solidity ^0.5.0;

contract TestContract {
  bytes32 public bytes32Data; //0
  struct TestStruct {
    uint256 var1;
  }

  mapping (uint => TestStruct) public testStructsData; //1
  address public addressData; //2
  uint48 public uint48Data; //2
  uint256[10] uintArrayData; //3

  constructor() public {
    uintArrayData[0] = 1;
  }

  function UpdateAllData() public {
    bytes32Data = "test data";
    addressData = 0x6c3abb37949F1e009bAF32fC91b69513769974D5;
    uint48Data = 2;
    uintArrayData[0] = 3;
    testStructsData[1].var1 = 4;
  }
}
*/


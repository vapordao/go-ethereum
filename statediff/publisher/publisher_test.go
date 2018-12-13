package publisher_test

import (
	"github.com/onsi/ginkgo"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/onsi/gomega"
	"os"
	"encoding/csv"
	"github.com/ethereum/go-ethereum/common"
	"math/rand"
	"math/big"
	"path/filepath"
	"strings"
	"strconv"
	p "github.com/ethereum/go-ethereum/statediff/publisher"
	"github.com/ethereum/go-ethereum/statediff/builder"
)

var _ = ginkgo.Describe("Publisher", func() {
	ginkgo.Context("default CSV publisher", func() {
		var (
			publisher p.Publisher
			err error
			config = statediff.Config{
				Path: "./test-",
			}
		)

		var (
			blockNumber = rand.Int63()
			blockHash = "0xfa40fbe2d98d98b3363a778d52f2bcd29d6790b9b3f3cab2b167fd12d3550f73"
			codeHash = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
			oldNonceValue = rand.Uint64()
			newNonceValue = oldNonceValue + 1
			oldBalanceValue = rand.Int63()
			newBalanceValue = oldBalanceValue - 1
			contractRoot = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
			storagePath = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
			oldStorage = "0x0"
			newStorage = "0x03"
			storage = map[string]builder.DiffString{storagePath: {
				NewValue: &newStorage,
				OldValue: &oldStorage,
			}}
			address = common.HexToAddress("0xaE9BEa628c4Ce503DcFD7E305CaB4e29E7476592")
			createdAccounts = map[common.Address]builder.AccountDiffEventual{address: {
				Nonce: builder.DiffUint64{
					NewValue: &newNonceValue,
					OldValue: &oldNonceValue,
				},
				Balance: builder.DiffBigInt{
					NewValue: big.NewInt(newBalanceValue),
					OldValue: big.NewInt(oldBalanceValue),
				},
				ContractRoot: builder.DiffString{
					NewValue: &contractRoot,
					OldValue: &contractRoot,
				},
				Code:     []byte("created account code"),
				CodeHash: codeHash,
				Storage:  storage,
			}}

			updatedAccounts = map[common.Address]builder.AccountDiffIncremental{address: {
				Nonce:        builder.DiffUint64{
					NewValue: &newNonceValue,
					OldValue: &oldNonceValue,
				},
				Balance:      builder.DiffBigInt{
					NewValue: big.NewInt(newBalanceValue),
					OldValue: big.NewInt(oldBalanceValue),
				},
				CodeHash:     codeHash,
				ContractRoot: builder.DiffString{
					NewValue: &contractRoot,
					OldValue: &contractRoot,
				},
				Storage: storage,
			}}

			deletedAccounts = map[common.Address]builder.AccountDiffEventual{address: {
				Nonce: builder.DiffUint64{
					NewValue: &newNonceValue,
					OldValue: &oldNonceValue,
				},
				Balance: builder.DiffBigInt{
					NewValue: big.NewInt(newBalanceValue),
					OldValue: big.NewInt(oldBalanceValue),
				},
				ContractRoot: builder.DiffString{
					NewValue: &contractRoot,
					OldValue: &contractRoot,
				},
				Code:     []byte("deleted account code"),
				CodeHash: codeHash,
				Storage:  storage,
			}}

			testStateDiff = builder.StateDiff{
				BlockNumber:     blockNumber,
				BlockHash:       common.HexToHash(blockHash),
				CreatedAccounts: createdAccounts,
				DeletedAccounts: deletedAccounts,
				UpdatedAccounts: updatedAccounts,
			}
		)

		var lines [][]string
		var file *os.File
		ginkgo.BeforeEach(func() {
			publisher, err = p.NewPublisher(config)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			_, err := publisher.PublishStateDiff(&testStateDiff)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			filePaths := getTestCSVFiles(".")
			file, err = os.Open(filePaths[0])
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			defer file.Close()

			lines, err = csv.NewReader(file).ReadAll()
		})

		ginkgo.AfterEach(func() {
			os.Remove(file.Name())
		})

		ginkgo.It("persists the column headers to a CSV file", func() {
			gomega.Expect(len(lines) > 1).To(gomega.BeTrue())
			gomega.Expect(lines[0]).To(gomega.Equal(p.Headers))
		})

		ginkgo.It("persists the created account diffs to a CSV file", func() {
			expectedCreatedAccountRow := []string{
				strconv.FormatInt(blockNumber, 10),
				blockHash,
				"created",
				"created account code",
				codeHash,
				strconv.FormatUint(oldNonceValue, 10),
				strconv.FormatUint(newNonceValue, 10),
				strconv.FormatInt(oldBalanceValue, 10),
				strconv.FormatInt(newBalanceValue, 10),
				contractRoot,
				contractRoot,
				storagePath,
			}

			gomega.Expect(len(lines) > 1).To(gomega.BeTrue())
			gomega.Expect(lines[1]).To(gomega.Equal(expectedCreatedAccountRow))
		})

		ginkgo.It("persists the updated account diffs to a CSV file", func() {
			expectedUpdatedAccountRow := []string{
				strconv.FormatInt(blockNumber, 10),
				blockHash,
				"updated",
				"",
				codeHash,
				strconv.FormatUint(oldNonceValue, 10),
				strconv.FormatUint(newNonceValue, 10),
				strconv.FormatInt(oldBalanceValue, 10),
				strconv.FormatInt(newBalanceValue, 10),
				contractRoot,
				contractRoot,
				storagePath,
			}

			gomega.Expect(len(lines) > 2).To(gomega.BeTrue())
			gomega.Expect(lines[2]).To(gomega.Equal(expectedUpdatedAccountRow))
		})

		ginkgo.It("persists the deleted account diffs to a CSV file", func() {
			expectedDeletedAccountRow := []string{
				strconv.FormatInt(blockNumber, 10),
				blockHash,
				"deleted",
				"deleted account code",
				codeHash,
				strconv.FormatUint(oldNonceValue, 10),
				strconv.FormatUint(newNonceValue, 10),
				strconv.FormatInt(oldBalanceValue, 10),
				strconv.FormatInt(newBalanceValue, 10),
				contractRoot,
				contractRoot,
				storagePath,
			}

			gomega.Expect(len(lines) > 3).To(gomega.BeTrue())
			gomega.Expect(lines[3]).To(gomega.Equal(expectedDeletedAccountRow))
		})
	})
})

func getTestCSVFiles(rootPath string) []string{
	var files []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(path, "test-") {
			files = append(files, path)
		}
		return nil
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return files
}

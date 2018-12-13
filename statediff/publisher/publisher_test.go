package publisher_test

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"os"
	"encoding/csv"
	"path/filepath"
	"strconv"
	p "github.com/ethereum/go-ethereum/statediff/publisher"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/statediff/builder"
	"github.com/ethereum/go-ethereum/statediff"
)

var _ = ginkgo.Describe("Publisher", func() {
	var (
		tempDir = os.TempDir()
		testFilePrefix = "test-statediff"
		publisher p.Publisher
		dir string
		err error
	)

	var expectedCreatedAccountRow = []string{
		strconv.FormatInt(testhelpers.BlockNumber, 10),
		testhelpers.BlockHash,
		"created",
		"created account code",
		testhelpers.CodeHash,
		strconv.FormatUint(testhelpers.OldNonceValue, 10),
		strconv.FormatUint(testhelpers.NewNonceValue, 10),
		strconv.FormatInt(testhelpers.OldBalanceValue, 10),
		strconv.FormatInt(testhelpers.NewBalanceValue, 10),
		testhelpers.ContractRoot,
		testhelpers.ContractRoot,
		testhelpers.StoragePath,
	}

	var expectedUpdatedAccountRow = []string{
		strconv.FormatInt(testhelpers.BlockNumber, 10),
		testhelpers.BlockHash,
		"updated",
		"",
		testhelpers.CodeHash,
		strconv.FormatUint(testhelpers.OldNonceValue, 10),
		strconv.FormatUint(testhelpers.NewNonceValue, 10),
		strconv.FormatInt(testhelpers.OldBalanceValue, 10),
		strconv.FormatInt(testhelpers.NewBalanceValue, 10),
		testhelpers.ContractRoot,
		testhelpers.ContractRoot,
		testhelpers.StoragePath,
	}

	var expectedDeletedAccountRow = []string{
		strconv.FormatInt(testhelpers.BlockNumber, 10),
		testhelpers.BlockHash,
		"deleted",
		"deleted account code",
		testhelpers.CodeHash,
		strconv.FormatUint(testhelpers.OldNonceValue, 10),
		strconv.FormatUint(testhelpers.NewNonceValue, 10),
		strconv.FormatInt(testhelpers.OldBalanceValue, 10),
		strconv.FormatInt(testhelpers.NewBalanceValue, 10),
		testhelpers.ContractRoot,
		testhelpers.ContractRoot,
		testhelpers.StoragePath,
	}

	ginkgo.BeforeEach(func() {
		dir, err = ioutil.TempDir(tempDir, testFilePrefix)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		config := statediff.Config{
			Path: dir,
			Mode: statediff.CSV,
		}
		publisher, err = p.NewPublisher(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		os.RemoveAll(dir)
	})

	ginkgo.It("persists the column headers to a CSV file", func() {
		_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		file, err := getTestDiffFile(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		lines, err := csv.NewReader(file).ReadAll()
		gomega.Expect(len(lines) > 1).To(gomega.BeTrue())
		gomega.Expect(lines[0]).To(gomega.Equal(p.Headers))
	})

	ginkgo.It("persists the created, upated and deleted account diffs to a CSV file", func() {
		_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		file, err := getTestDiffFile(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		lines, err := csv.NewReader(file).ReadAll()
		gomega.Expect(len(lines) > 3).To(gomega.BeTrue())
		gomega.Expect(lines[1]).To(gomega.Equal(expectedCreatedAccountRow))
		gomega.Expect(lines[2]).To(gomega.Equal(expectedUpdatedAccountRow))
		gomega.Expect(lines[3]).To(gomega.Equal(expectedDeletedAccountRow))
	})

	ginkgo.It("creates an empty CSV when there is no diff", func() {
		emptyDiff := builder.StateDiff{}
		_, err = publisher.PublishStateDiff(&emptyDiff)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		file, err := getTestDiffFile(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		lines, err := csv.NewReader(file).ReadAll()
		gomega.Expect(len(lines)).To(gomega.Equal(1))
	})

	ginkgo.It("defaults to publishing state diffs to a CSV file when no mode is configured", func() {
		config := statediff.Config{Path: dir}
		publisher, err = p.NewPublisher(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		file, err := getTestDiffFile(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		lines, err := csv.NewReader(file).ReadAll()
		gomega.Expect(len(lines)).To(gomega.Equal(4))
		gomega.Expect(lines[0]).To(gomega.Equal(p.Headers))
	})

	ginkgo.FIt("defaults to publishing CSV files in the current directory when no path is configured", func() {
		config := statediff.Config{}
		publisher, err = p.NewPublisher(config)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		err := os.Chdir(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		file, err := getTestDiffFile(dir)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		lines, err := csv.NewReader(file).ReadAll()
		gomega.Expect(len(lines)).To(gomega.Equal(4))
		gomega.Expect(lines[0]).To(gomega.Equal(p.Headers))
	})
})

func getTestDiffFile(dir string) (*os.File, error) {
	files, err := ioutil.ReadDir(dir)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(len(files) > 0).To(gomega.BeTrue())

	fileName := files[0].Name()
	filePath := filepath.Join(dir, fileName)

	return os.Open(filePath)
}

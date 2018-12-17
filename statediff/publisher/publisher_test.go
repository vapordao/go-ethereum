package publisher_test

import (
	"testing"
	"os"
	"strconv"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
	p "github.com/ethereum/go-ethereum/statediff/publisher"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/statediff"
	"encoding/csv"
	"path/filepath"
	"bytes"
	"reflect"
	"github.com/ethereum/go-ethereum/statediff/builder"
	"github.com/pkg/errors"
)

var (
	tempDir        = os.TempDir()
	testFilePrefix = "test-statediff"
	publisher      p.Publisher
	dir            string
	err            error
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



func TestPublisher(t *testing.T) {
	dir, err = ioutil.TempDir(tempDir, testFilePrefix)
	if err != nil {
		t.Error(err)
	}
	config := statediff.Config{
		Path: dir,
		Mode: statediff.CSV,
	}
	publisher, err = p.NewPublisher(config)
	if err != nil {
		t.Error(err)
	}

	type Test func(t *testing.T)

	var tests = []Test{testColumnHeaders,
		testAccountDiffs,
		testWhenNoDiff,
		testDefaultPublisher,
		testDefaultDirectory,
	}

	for _, test := range tests {
		test(t)
		removeFilesFromDir(dir, t)
	}
}

func removeFilesFromDir(dir string, t *testing.T) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		t.Error()
	}

	for _, file := range files {
		err = os.RemoveAll(file)
		if err !=nil {
			t.Error()
		}
	}
}

func testColumnHeaders(t *testing.T) {
	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Error(err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Error(err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Error(err)
	}
	if len(lines) <= 1 { t.Error() }

	if !equals(lines[0], p.Headers) { t.Error() }
}

func testAccountDiffs(t *testing.T) {
	// it persists the created, updated and deleted account diffs to a CSV file
	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Error(err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Error(err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Error(err)
	}
	if len(lines) <= 3 { t.Error() }
	if !equals(lines[1], expectedCreatedAccountRow) { t.Error() }
	if !equals(lines[2], expectedUpdatedAccountRow) { t.Error()}
	if !equals(lines[3], expectedDeletedAccountRow) { t.Error()}
}

func testWhenNoDiff(t *testing.T) {
	//it creates an empty CSV when there is no diff", func() {
	emptyDiff := builder.StateDiff{}
	_, err = publisher.PublishStateDiff(&emptyDiff)
	if err != nil {
		t.Error(err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Error(err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Error(err)
	}
	if !equals(len(lines), 1) { t.Error() }
}

func testDefaultPublisher(t *testing.T) {
	//it defaults to publishing state diffs to a CSV file when no mode is configured
	config := statediff.Config{Path: dir}
	publisher, err = p.NewPublisher(config)
	if err != nil { t.Error(err) }

	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil { t.Error(err) }

	file, err := getTestDiffFile(dir)
	if err != nil { t.Error(err) }

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil { t.Error(err) }
	if !equals(len(lines), 4) { t.Error()}
	if !equals(lines[0],p.Headers) { t.Error()}
}

func testDefaultDirectory(t *testing.T) {
	//it defaults to publishing CSV files in the current directory when no path is configured
	config := statediff.Config{}
	publisher, err = p.NewPublisher(config)
	if err != nil { t.Error(err) }

	err := os.Chdir(dir)
	if err != nil { t.Error(err) }

	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil { t.Error(err) }

	file, err := getTestDiffFile(dir)
	if err != nil { t.Error(err) }

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil { t.Error(err) }
	if !equals(len(lines), 4) { t.Error() }
	if !equals(lines[0], p.Headers) { t.Error() }
}

func getTestDiffFile(dir string) (*os.File, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil { return nil, err }
	if len(files) == 0 { return nil, errors.New("There are 0 files.") }

	fileName := files[0].Name()
	filePath := filepath.Join(dir, fileName)

	return os.Open(filePath)
}

func equals(actual, expected interface{}) (success bool) {
	if actualByteSlice, ok := actual.([]byte); ok {
		if expectedByteSlice, ok := expected.([]byte); ok {
			return bytes.Equal(actualByteSlice, expectedByteSlice)
		}
	}

	return reflect.DeepEqual(actual, expected)
}

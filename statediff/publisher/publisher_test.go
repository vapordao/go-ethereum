package publisher_test

import (
	"bytes"
	"encoding/csv"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/ethereum/go-ethereum/statediff/builder"
	p "github.com/ethereum/go-ethereum/statediff/publisher"
	"github.com/ethereum/go-ethereum/statediff/testhelpers"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

var (
	tempDir        = os.TempDir()
	testFilePrefix = "test-statediff"
	publisher      p.Publisher
	dir            string
	err            error
    testFailedFormatString = "Test failed: %s, %+v"
)

var expectedCreatedAccountRow = []string{
	strconv.FormatInt(testhelpers.BlockNumber, 10),
	testhelpers.BlockHash,
	"created",
	"created account code",
	testhelpers.CodeHash,
	strconv.FormatUint(testhelpers.NewNonceValue, 10),
	strconv.FormatInt(testhelpers.NewBalanceValue, 10),
	testhelpers.ContractRoot,
	testhelpers.StoragePath,
}

var expectedUpdatedAccountRow = []string{
	strconv.FormatInt(testhelpers.BlockNumber, 10),
	testhelpers.BlockHash,
	"updated",
	"",
	testhelpers.CodeHash,
	strconv.FormatUint(testhelpers.NewNonceValue, 10),
	strconv.FormatInt(testhelpers.NewBalanceValue, 10),
	testhelpers.ContractRoot,
	testhelpers.StoragePath,
}

var expectedDeletedAccountRow = []string{
	strconv.FormatInt(testhelpers.BlockNumber, 10),
	testhelpers.BlockHash,
	"deleted",
	"deleted account code",
	testhelpers.CodeHash,
	strconv.FormatUint(testhelpers.NewNonceValue, 10),
	strconv.FormatInt(testhelpers.NewBalanceValue, 10),
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

	var tests = []Test{
		testColumnHeaders,
		testAccountDiffs,
		testWhenNoDiff,
		testDefaultPublisher,
		testDefaultDirectory,
	}

	for _, test := range tests {
		test(t)
		err := removeFilesFromDir(dir)
		if err != nil {
			t.Error("Error removing files from temp dir: %s", dir)
		}
	}
}

func removeFilesFromDir(dir string,) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}

	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func testColumnHeaders(t *testing.T) {
	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if len(lines) < 1 {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[0], p.Headers) {
		t.Error()
	}
}

func testAccountDiffs(t *testing.T) {
	// it persists the created, updated and deleted account diffs to a CSV file
	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if len(lines) <= 3 {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[1], expectedCreatedAccountRow) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[2], expectedUpdatedAccountRow) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[3], expectedDeletedAccountRow) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
}

func testWhenNoDiff(t *testing.T) {
	//it creates an empty CSV when there is no diff
	emptyDiff := builder.StateDiff{}
	_, err = publisher.PublishStateDiff(&emptyDiff)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	if !equals(len(lines), 1) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
}

func testDefaultPublisher(t *testing.T) {
	//it defaults to publishing state diffs to a CSV file when no mode is configured
	config := statediff.Config{Path: dir}
	publisher, err = p.NewPublisher(config)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(len(lines), 4) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[0], p.Headers) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
}

func testDefaultDirectory(t *testing.T) {
	//it defaults to publishing CSV files in the current directory when no path is configured
	config := statediff.Config{}
	publisher, err = p.NewPublisher(config)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	err := os.Chdir(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	_, err = publisher.PublishStateDiff(&testhelpers.TestStateDiff)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	file, err := getTestDiffFile(dir)
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(len(lines), 4) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
	if !equals(lines[0], p.Headers) {
		t.Errorf(testFailedFormatString, t.Name(), err)
	}
}

func getTestDiffFile(dir string) (*os.File, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("There are 0 files.")
	}

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

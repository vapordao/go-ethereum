// Copyright 2015 The go-ethereum Authors
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

// Contains a batch of utility type declarations used by the tests. As the node
// operates on unique types, a lot of them are needed to check various features.

package publisher

import (
	"os"
	"encoding/csv"
	"time"
	"strconv"
	"strings"
	"github.com/ethereum/go-ethereum/statediff/builder"
	"github.com/ethereum/go-ethereum/statediff"
)

type Publisher interface {
	PublishStateDiff(sd *builder.StateDiff) (string, error)
}

type publisher struct {
	Config statediff.Config
}

var (
	Headers = []string{
		"blockNumber", "blockHash", "accountAction",
		"code", "codeHash",
		"oldNonceValue", "newNonceValue",
		"oldBalanceValue", "newBalanceValue",
		"oldContractRoot", "newContractRoot",
		"storageDiffPaths",
	}

	timeStampFormat = "20060102150405.00000"
	deletedAccountAction = "deleted"
	createdAccountAction = "created"
	updatedAccountAction = "updated"
)

func NewPublisher(config statediff.Config) (*publisher, error) {
	return &publisher{
		Config: config,
	}, nil
}

func (p *publisher) PublishStateDiff(sd *builder.StateDiff) (string, error) {
	switch p.Config.Mode {
	case statediff.CSV:
		return "", p.publishStateDiffToCSV(*sd)
	default:
		return "", p.publishStateDiffToCSV(*sd)
	}
}

func (p *publisher) publishStateDiffToCSV(sd builder.StateDiff) error {
	now := time.Now()
	timeStamp := now.Format(timeStampFormat)
	filePath := p.Config.Path + timeStamp + ".csv"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var data [][]string
	data = append(data, Headers)
	for _, row := range accumulateCreatedAccountRows(sd) {
		data = append(data, row)
	}
	for _, row := range accumulateUpdatedAccountRows(sd) {
		data = append(data, row)
	}

	for _, row := range accumulateDeletedAccountRows(sd) {
		data = append(data, row)
	}

	for _, value := range data{
		err := writer.Write(value)
		if err != nil {
			return err
		}
	}

	return nil
}

func accumulateUpdatedAccountRows(sd builder.StateDiff) [][]string {
	var updatedAccountRows [][]string
	for _, accountDiff := range sd.UpdatedAccounts {
		formattedAccountData := formatAccountDiffIncremental(accountDiff, sd, updatedAccountAction)

		updatedAccountRows = append(updatedAccountRows, formattedAccountData)
	}

	return updatedAccountRows
}

func accumulateDeletedAccountRows(sd builder.StateDiff) [][]string {
	var deletedAccountRows [][]string
	for _, accountDiff := range sd.DeletedAccounts {
		formattedAccountData := formatAccountDiffEventual(accountDiff, sd, deletedAccountAction)

		deletedAccountRows = append(deletedAccountRows, formattedAccountData)
	}

	return deletedAccountRows
}

func accumulateCreatedAccountRows(sd builder.StateDiff) [][]string {
	var createdAccountRows [][]string
	for _, accountDiff := range sd.CreatedAccounts {
		formattedAccountData := formatAccountDiffEventual(accountDiff, sd, createdAccountAction)

		createdAccountRows = append(createdAccountRows, formattedAccountData)
	}

	return createdAccountRows
}

func formatAccountDiffEventual(accountDiff builder.AccountDiffEventual, sd builder.StateDiff, accountAction string) []string {
	oldContractRoot := accountDiff.ContractRoot.OldValue
	newContractRoot := accountDiff.ContractRoot.NewValue
	var storageDiffPaths []string
	for k := range accountDiff.Storage {
		storageDiffPaths = append(storageDiffPaths, k)
	}
	formattedAccountData := []string{
		strconv.FormatInt(sd.BlockNumber, 10),
		sd.BlockHash.String(),
		accountAction,
		string(accountDiff.Code),
		accountDiff.CodeHash,
		strconv.FormatUint(*accountDiff.Nonce.OldValue, 10),
		strconv.FormatUint(*accountDiff.Nonce.NewValue, 10),
		accountDiff.Balance.OldValue.String(),
		accountDiff.Balance.NewValue.String(),
		*oldContractRoot,
		*newContractRoot,
		strings.Join(storageDiffPaths, ","),
	}
	return formattedAccountData
}

func formatAccountDiffIncremental(accountDiff builder.AccountDiffIncremental, sd builder.StateDiff, accountAction string) []string {
	oldContractRoot := accountDiff.ContractRoot.OldValue
	newContractRoot := accountDiff.ContractRoot.NewValue
	var storageDiffPaths []string
	for k := range accountDiff.Storage {
		storageDiffPaths = append(storageDiffPaths, k)
	}
	formattedAccountData := []string{
		strconv.FormatInt(sd.BlockNumber, 10),
		sd.BlockHash.String(),
		accountAction,
		"",
		accountDiff.CodeHash,
		strconv.FormatUint(*accountDiff.Nonce.OldValue, 10),
		strconv.FormatUint(*accountDiff.Nonce.NewValue, 10),
		accountDiff.Balance.OldValue.String(),
		accountDiff.Balance.NewValue.String(),
		*oldContractRoot,
		*newContractRoot,
		strings.Join(storageDiffPaths, ","),
	}
	return formattedAccountData
}


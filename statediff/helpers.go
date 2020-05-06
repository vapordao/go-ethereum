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

// Contains a batch of utility type declarations used by the tests. As the node
// operates on unique types, a lot of them are needed to check various features.

package statediff

import (
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie"
)

func sortKeys(data AccountsMap) []string {
	var keys []string
	for key := range data {
		keys = append(keys, key.Hex())
	}
	sort.Strings(keys)

	return keys
}

// bytesToNiblePath converts the byte representation of a path to its string representation
func bytesToNiblePath(path []byte) string {
	if hasTerm(path) {
		path = path[:len(path)-1]
	}
	nibblePath := ""
	for i, v := range common.ToHex(path) {
		if i%2 == 0 && i > 1 {
			continue
		}
		nibblePath = nibblePath + string(v)
	}

	return nibblePath
}

// findIntersection finds the set of strings from both arrays that are equivalent (same key as same index)
// this is used to find which keys have been both "deleted" and "created" i.e. they were updated
func findIntersection(created, deleted AccountsMap) []string {
	var keys []string
	for key := range created {
		_, ok := deleted[key]
		if ok {
			keys = append(keys, key.Hex())
		}
	}

	for key, _ := range deleted {
		_, ok := created[key]
		if ok {
			keys = append(keys, key.Hex())
		}
	}
	uniqueKeys := make(map[string]bool)
	var uniqueKeysList []string
	for _, entry := range keys {
		if _, ok := uniqueKeys[entry]; !ok {
			uniqueKeys[entry] = true
			uniqueKeysList = append(uniqueKeysList, entry)
		}
	}

	return uniqueKeysList
}

// pathToStr converts the NodeIterator path to a string representation
func pathToStr(it trie.NodeIterator) string {
	return bytesToNiblePath(it.Path())
}

// hasTerm returns whether a hex key has the terminator flag.
func hasTerm(s []byte) bool {
	return len(s) > 0 && s[len(s)-1] == 16
}

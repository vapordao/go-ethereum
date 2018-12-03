package statediff_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStatediff(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Statediff Suite")
}

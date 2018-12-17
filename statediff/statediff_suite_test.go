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

//convert this over to use built in golang library
//only save the new value, and have a pointer to the old value - not sure how this pointer will work for the CSV version

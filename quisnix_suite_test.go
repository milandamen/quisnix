package quisnix_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestQuisnix(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Quisnix Suite")
}

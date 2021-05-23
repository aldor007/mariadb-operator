package controllers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestMariaDBControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mariadb Controller Spec")
}

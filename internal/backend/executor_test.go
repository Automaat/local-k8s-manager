package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

var _ = Describe("Executor", func() {
	Describe("IsCommandAvailable", func() {
		It("should return true for existing commands", func() {
			result := backend.IsCommandAvailable("ls")
			Expect(result).To(BeTrue())
		})

		It("should return false for non-existing commands", func() {
			result := backend.IsCommandAvailable("this-command-does-not-exist-12345")
			Expect(result).To(BeFalse())
		})
	})

	Describe("Exec", func() {
		It("should execute commands successfully", func() {
			output, err := backend.Exec("echo", "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())
		})

		It("should return error for failing commands", func() {
			_, err := backend.Exec("ls", "/this-path-does-not-exist-12345")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("DefaultExecutor", func() {
		var executor *backend.DefaultExecutor

		BeforeEach(func() {
			executor = &backend.DefaultExecutor{}
		})

		It("should execute commands", func() {
			output, err := executor.Exec("echo", "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(output).NotTo(BeEmpty())
		})

		It("should check command availability", func() {
			result := executor.IsCommandAvailable("ls")
			Expect(result).To(BeTrue())
		})
	})
})

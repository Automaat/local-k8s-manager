package backend_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/automaat/local-k8s-manager/internal/backend"
)

var _ = Describe("Interface", func() {
	Describe("CreateOptions", func() {
		It("should store workers count", func() {
			opts := backend.CreateOptions{Workers: 3}
			Expect(opts.Workers).To(Equal(3))
		})
	})
})

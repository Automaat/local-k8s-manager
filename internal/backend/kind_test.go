package backend_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
)

var _ = Describe("KindProvider", func() {
	var (
		provider     *backend.KindProvider
		mockExecutor *MockExecutor
	)

	BeforeEach(func() {
		provider = backend.NewKindProvider()
		mockExecutor = &MockExecutor{}
		backend.SetExecutor(mockExecutor)
	})

	AfterEach(func() {
		backend.SetExecutor(&backend.DefaultExecutor{})
	})

	Describe("Name", func() {
		It("should return kind", func() {
			Expect(provider.Name()).To(Equal("kind"))
		})
	})

	Describe("IsInstalled", func() {
		Context("when kind is available", func() {
			It("should return true", func() {
				mockExecutor.IsCommandAvailableFunc = func(name string) bool {
					return name == "kind"
				}
				Expect(provider.IsInstalled()).To(BeTrue())
			})
		})

		Context("when kind is not available", func() {
			It("should return false", func() {
				mockExecutor.IsCommandAvailableFunc = func(name string) bool {
					return false
				}
				Expect(provider.IsInstalled()).To(BeFalse())
			})
		})
	})

	Describe("List", func() {
		Context("when command succeeds", func() {
			It("should return list of clusters", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte("cluster1\ncluster2\n"), nil
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(2))
				Expect(clusters[0].Name).To(Equal("cluster1"))
				Expect(clusters[1].Name).To(Equal("cluster2"))
				Expect(clusters[0].Status).To(Equal(models.StatusRunning))
			})
		})

		Context("when no clusters exist", func() {
			It("should return empty list when command returns error with empty output", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte(""), errors.New("no clusters")
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(0))
			})

			It("should return empty list when command returns error with 'No kind clusters found' message", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte("No kind clusters found."), errors.New("exit status 1")
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(0))
			})

			It("should return empty list when command succeeds with 'No kind clusters found' message", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte("No kind clusters found."), nil
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(0))
			})
		})

		Context("when command fails with real error", func() {
			It("should return error", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte("some other error"), errors.New("command failed")
				}

				_, err := provider.List()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Create", func() {
		It("should create cluster", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("success"), nil
			}

			err := provider.Create("test-cluster", backend.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
		})

		It("should return Docker error when Docker daemon is not running", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("Cannot connect to the Docker daemon"), errors.New("exit status 1")
			}

			err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Docker is not running"))
		})

		It("should return command output on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("cluster name already exists"), errors.New("exit status 1")
			}

			err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster name already exists"))
		})
	})

	Describe("Delete", func() {
		It("should delete cluster", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("success"), nil
			}

			err := provider.Delete("test-cluster")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Delete("test")
			Expect(err).To(HaveOccurred())
		})

		It("should return Docker error when Docker daemon is not running", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("Docker daemon is not running"), errors.New("exit status 1")
			}

			err := provider.Delete("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Docker is not running"))
		})

		It("should return command output on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("cluster not found"), errors.New("exit status 1")
			}

			err := provider.Delete("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster not found"))
		})
	})

	Describe("Start", func() {
		It("should return unsupported error", func() {
			err := provider.Start("test-cluster")
			Expect(err).To(MatchError(ContainSubstring("does not support")))
		})
	})

	Describe("Stop", func() {
		It("should return unsupported error", func() {
			err := provider.Stop("test-cluster")
			Expect(err).To(MatchError(ContainSubstring("does not support")))
		})
	})
})

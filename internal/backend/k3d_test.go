package backend_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
)

var _ = Describe("K3dProvider", func() {
	var (
		provider     *backend.K3dProvider
		mockExecutor *MockExecutor
	)

	BeforeEach(func() {
		provider = backend.NewK3dProvider()
		mockExecutor = &MockExecutor{}
		backend.SetExecutor(mockExecutor)
	})

	AfterEach(func() {
		backend.SetExecutor(&backend.DefaultExecutor{})
	})

	Describe("Name", func() {
		It("should return k3d", func() {
			Expect(provider.Name()).To(Equal("k3d"))
		})
	})

	Describe("IsInstalled", func() {
		Context("when k3d is available", func() {
			It("should return true", func() {
				mockExecutor.IsCommandAvailableFunc = func(name string) bool {
					return name == "k3d"
				}
				Expect(provider.IsInstalled()).To(BeTrue())
			})
		})

		Context("when k3d is not available", func() {
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
					clusters := []backend.K3dCluster{
						{
							Name: "test-cluster",
							Nodes: []backend.K3dNode{
								{Name: "node1", State: backend.K3dNodeState{Running: true}},
								{Name: "node2", State: backend.K3dNodeState{Running: true}},
							},
							ServerCount: 1,
						},
					}
					return json.Marshal(clusters)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Name).To(Equal("test-cluster"))
				Expect(clusters[0].Status).To(Equal(models.StatusRunning))
				Expect(clusters[0].Nodes).To(Equal(2))
			})
		})

		Context("when some nodes are stopped", func() {
			It("should return unknown status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					clusters := []backend.K3dCluster{
						{
							Name: "partial-cluster",
							Nodes: []backend.K3dNode{
								{Name: "node1", State: backend.K3dNodeState{Running: true}},
								{Name: "node2", State: backend.K3dNodeState{Running: false}},
							},
						},
					}
					return json.Marshal(clusters)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters[0].Status).To(Equal(models.StatusUnknown))
			})
		})

		Context("when all nodes are stopped", func() {
			It("should return stopped status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					clusters := []backend.K3dCluster{
						{
							Name: "stopped-cluster",
							Nodes: []backend.K3dNode{
								{Name: "node1", State: backend.K3dNodeState{Running: false}},
								{Name: "node2", State: backend.K3dNodeState{Running: false}},
							},
						},
					}
					return json.Marshal(clusters)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters[0].Status).To(Equal(models.StatusStopped))
			})
		})

		Context("when command fails", func() {
			It("should return error", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return nil, errors.New("command failed")
				}

				_, err := provider.List()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when JSON is invalid", func() {
			It("should return error", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte("invalid json"), nil
				}

				_, err := provider.List()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Create", func() {
		It("should create cluster without workers", func() {
			var capturedArgs []string
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				capturedArgs = args
				return []byte("success"), nil
			}

			err := provider.Create("test-cluster", backend.CreateOptions{Workers: 0})
			Expect(err).NotTo(HaveOccurred())
			Expect(capturedArgs).To(ContainElement("test-cluster"))
			Expect(capturedArgs).NotTo(ContainElement("--agents"))
		})

		It("should create cluster with workers", func() {
			var capturedArgs []string
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				capturedArgs = args
				return []byte("success"), nil
			}

			err := provider.Create("test-cluster", backend.CreateOptions{Workers: 2})
			Expect(err).NotTo(HaveOccurred())
			Expect(capturedArgs).To(ContainElements("--agents", "2"))
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Delete", func() {
		It("should delete cluster", func() {
			var capturedArgs []string
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				capturedArgs = args
				return []byte("success"), nil
			}

			err := provider.Delete("test-cluster")
			Expect(err).NotTo(HaveOccurred())
			Expect(capturedArgs).To(ContainElement("test-cluster"))
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Delete("test")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Start", func() {
		It("should start cluster", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("success"), nil
			}

			err := provider.Start("test-cluster")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Start("test")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Stop", func() {
		It("should stop cluster", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("success"), nil
			}

			err := provider.Stop("test-cluster")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			err := provider.Stop("test")
			Expect(err).To(HaveOccurred())
		})
	})
})

package backend_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/models"
)

var _ = Describe("MinikubeProvider", func() {
	var (
		provider     *backend.MinikubeProvider
		mockExecutor *MockExecutor
	)

	BeforeEach(func() {
		provider = backend.NewMinikubeProvider()
		mockExecutor = &MockExecutor{}
		backend.SetExecutor(mockExecutor)
	})

	AfterEach(func() {
		backend.SetExecutor(&backend.DefaultExecutor{})
	})

	Describe("Name", func() {
		It("should return minikube", func() {
			Expect(provider.Name()).To(Equal("minikube"))
		})
	})

	Describe("IsInstalled", func() {
		Context("when minikube is available", func() {
			It("should return true", func() {
				mockExecutor.IsCommandAvailableFunc = func(name string) bool {
					return name == "minikube"
				}
				Expect(provider.IsInstalled()).To(BeTrue())
			})
		})

		Context("when minikube is not available", func() {
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
			It("should return list of clusters with Running status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					response := map[string]interface{}{
						"valid": []map[string]interface{}{
							{
								"Name":   "test-cluster",
								"Status": "Running",
								"Config": map[string]interface{}{
									"Nodes": []map[string]interface{}{
										{"Name": "node1"},
										{"Name": "node2"},
									},
								},
							},
						},
					}
					return json.Marshal(response)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Name).To(Equal("test-cluster"))
				Expect(clusters[0].Status).To(Equal(models.StatusRunning))
				Expect(clusters[0].Nodes).To(Equal(2))
			})

			It("should return list of clusters with OK status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					response := map[string]interface{}{
						"valid": []map[string]interface{}{
							{
								"Name":   "test-cluster",
								"Status": "OK",
								"Config": map[string]interface{}{
									"Nodes": []map[string]interface{}{
										{"Name": "node1"},
										{"Name": "node2"},
									},
								},
							},
						},
					}
					return json.Marshal(response)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Name).To(Equal("test-cluster"))
				Expect(clusters[0].Status).To(Equal(models.StatusRunning))
				Expect(clusters[0].Nodes).To(Equal(2))
			})

			It("should handle stderr noise before JSON", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					response := map[string]interface{}{
						"valid": []map[string]interface{}{
							{
								"Name":   "test-cluster",
								"Status": "OK",
								"Config": map[string]interface{}{
									"Nodes": []map[string]interface{}{
										{"Name": "node1"},
									},
								},
							},
						},
					}
					jsonData, _ := json.Marshal(response)
					// Prepend stderr error messages like minikube does
					withStderr := append([]byte("E1230 08:29:17.389884 status.go:458] some error\n"), jsonData...)
					return withStderr, nil
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Name).To(Equal("test-cluster"))
				Expect(clusters[0].Status).To(Equal(models.StatusRunning))
			})

			It("should return list with Stopped status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					response := map[string]interface{}{
						"valid": []map[string]interface{}{
							{
								"Name":   "stopped-cluster",
								"Status": "Stopped",
								"Config": map[string]interface{}{
									"Nodes": []map[string]interface{}{
										{"Name": "node1"},
									},
								},
							},
						},
					}
					return json.Marshal(response)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Status).To(Equal(models.StatusStopped))
			})

			It("should return Unknown status for unrecognized status", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					response := map[string]interface{}{
						"valid": []map[string]interface{}{
							{
								"Name":   "unknown-cluster",
								"Status": "SomeWeirdStatus",
								"Config": map[string]interface{}{
									"Nodes": []map[string]interface{}{
										{"Name": "node1"},
									},
								},
							},
						},
					}
					return json.Marshal(response)
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(1))
				Expect(clusters[0].Status).To(Equal(models.StatusUnknown))
			})
		})

		Context("when no clusters exist", func() {
			It("should return empty list", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					return []byte(""), errors.New("no profiles")
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(0))
			})
		})

		Context("when command fails with output", func() {
			It("should try to parse output", func() {
				mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
					// Even with error, if there's valid JSON, it should parse
					response := map[string]interface{}{
						"valid": []map[string]interface{}{},
					}
					data, _ := json.Marshal(response)
					return data, errors.New("some error")
				}

				clusters, err := provider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).To(HaveLen(0))
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

			output, err := provider.Create("test-cluster", backend.CreateOptions{Workers: 0})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("success"))
			Expect(capturedArgs).To(ContainElements("start", "--profile", "test-cluster"))
			Expect(capturedArgs).NotTo(ContainElement("--nodes"))
		})

		It("should create cluster with workers", func() {
			var capturedArgs []string
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				capturedArgs = args
				return []byte("success"), nil
			}

			output, err := provider.Create("test-cluster", backend.CreateOptions{Workers: 2})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("success"))
			Expect(capturedArgs).To(ContainElements("--nodes", "3")) // 2 workers + 1 control plane
		})

		It("should return error on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("failed")
			}

			output, err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
			Expect(output).To(Equal(""))
		})

		It("should return Docker error when Docker daemon is not running", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("Cannot connect to the Docker daemon"), errors.New("exit status 1")
			}

			output, err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
			Expect(output).To(Equal("Cannot connect to the Docker daemon"))
			Expect(err.Error()).To(ContainSubstring("Docker is not running"))
		})

		It("should return command output on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("profile already exists"), errors.New("exit status 1")
			}

			output, err := provider.Create("test", backend.CreateOptions{})
			Expect(err).To(HaveOccurred())
			Expect(output).To(Equal("profile already exists"))
			Expect(err.Error()).To(ContainSubstring("profile already exists"))
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
			Expect(capturedArgs).To(ContainElements("delete", "-p", "test-cluster"))
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
				return []byte("profile not found"), errors.New("exit status 1")
			}

			err := provider.Delete("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("profile not found"))
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

		It("should return Docker error when Docker daemon is not running", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("Is the docker daemon running?"), errors.New("exit status 1")
			}

			err := provider.Start("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Docker is not running"))
		})

		It("should return command output on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("cluster already running"), errors.New("exit status 1")
			}

			err := provider.Start("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster already running"))
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

		It("should return Docker error when Docker daemon is not running", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("docker: not found"), errors.New("exit status 1")
			}

			err := provider.Stop("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Docker is not running"))
		})

		It("should return command output on failure", func() {
			mockExecutor.ExecFunc = func(name string, args ...string) ([]byte, error) {
				return []byte("cluster already stopped"), errors.New("exit status 1")
			}

			err := provider.Stop("test")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cluster already stopped"))
		})
	})
})

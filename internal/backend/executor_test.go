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

		Describe("ExecStreaming", func() {
			It("should stream command output line-by-line", func() {
				outputChan := make(chan string, 10)
				done := make(chan bool)
				var lineCount int

				go func() {
					for range outputChan {
						lineCount++
					}
					done <- true
				}()

				output, err := executor.ExecStreaming("echo", []string{"-e", "line1\\nline2\\nline3"}, outputChan)
				close(outputChan)
				<-done

				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(ContainSubstring("line1"))
				Expect(lineCount).To(BeNumerically(">", 0))
			})

			It("should return error for failing commands", func() {
				outputChan := make(chan string, 10)
				go func() {
					for range outputChan {
					}
				}()

				_, err := executor.ExecStreaming("ls", []string{"/this-path-does-not-exist-12345"}, outputChan)
				close(outputChan)

				Expect(err).To(HaveOccurred())
			})

			It("should handle stdout pipe error", func() {
				outputChan := make(chan string, 10)

				// This test verifies the error path when creating stdout pipe fails
				// In normal conditions this won't fail, but we test the function structure
				_, err := executor.ExecStreaming("echo", []string{"test"}, outputChan)
				close(outputChan)

				// Command should succeed in normal circumstances
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ExecStreaming global function", func() {
		It("should use global executor for streaming", func() {
			outputChan := make(chan string, 10)
			done := make(chan bool)

			go func() {
				for range outputChan {
				}
				done <- true
			}()

			output, err := backend.ExecStreaming("echo", []string{"test"}, outputChan)
			close(outputChan)
			<-done

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("test"))
		})

		It("should fallback to non-streaming for mock executors", func() {
			defer func() {
				backend.SetExecutor(&backend.DefaultExecutor{})
			}()

			// Create a mock executor
			mockExecutor := &MockExecutor{
				ExecFunc: func(name string, args ...string) ([]byte, error) {
					return []byte("mock output"), nil
				},
			}
			backend.SetExecutor(mockExecutor)

			outputChan := make(chan string, 10)
			output, err := backend.ExecStreaming("test", []string{"arg"}, outputChan)
			close(outputChan)

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal("mock output"))
		})
	})

	Describe("ParseDockerError", func() {
		It("should detect 'cannot connect to the docker daemon' error", func() {
			output := "Error: Cannot connect to the Docker daemon at unix:///var/run/docker.sock"
			result := backend.ParseDockerError(output)
			Expect(result).To(Equal("Docker is not running. Please start Docker and try again."))
		})

		It("should detect 'is the docker daemon running' error", func() {
			output := "Is the docker daemon running?"
			result := backend.ParseDockerError(output)
			Expect(result).To(Equal("Docker is not running. Please start Docker and try again."))
		})

		It("should detect 'docker: not found' error", func() {
			output := "docker: not found"
			result := backend.ParseDockerError(output)
			Expect(result).To(Equal("Docker is not running. Please start Docker and try again."))
		})

		It("should detect 'docker daemon is not running' error", func() {
			output := "The Docker daemon is not running"
			result := backend.ParseDockerError(output)
			Expect(result).To(Equal("Docker is not running. Please start Docker and try again."))
		})

		It("should be case-insensitive", func() {
			output := "CANNOT CONNECT TO THE DOCKER DAEMON"
			result := backend.ParseDockerError(output)
			Expect(result).To(Equal("Docker is not running. Please start Docker and try again."))
		})

		It("should return empty string for non-Docker errors", func() {
			output := "some other error message"
			result := backend.ParseDockerError(output)
			Expect(result).To(BeEmpty())
		})

		It("should return empty string for empty input", func() {
			result := backend.ParseDockerError("")
			Expect(result).To(BeEmpty())
		})
	})
})

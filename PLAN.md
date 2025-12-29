# Local K8s Cluster Manager - Implementation Plan

TUI app for managing k3d/kind/minikube clusters - lifecycle operations only.

## Tech Stack

- **Language:** Go 1.23
- **TUI Framework:** Bubbletea (Elm architecture)
- **UI Components:** Bubbles (table, viewport)
- **Styling:** Lipgloss
- **Tools:** mise for dependency management
- **CI/CD:** GitHub Actions

## Decisions

- **Binary name:** `lkm`
- **Create options:** CLI form only (no config files)
- **Auto-install:** No - warn if tools missing
- **kubectl integration:** No - pure cluster lifecycle
- **Future providers:** Maybe later (k3s, k0s, microk8s)
- **Logging:** Yes - log operations to file

## Project Structure

```
.mise.toml                           # Tool versions and tasks
.github/workflows/ci.yml             # CI/CD pipeline
.golangci.yml                        # Linter config
cmd/lkm/main.go                      # Entry point
internal/
  backend/
    interface.go                     # Provider interface
    k3d.go                           # k3d CLI wrapper
    kind.go                          # kind CLI wrapper
    minikube.go                      # minikube CLI wrapper
    executor.go                      # Shell exec helpers
  models/
    cluster.go                       # Cluster struct
  tui/
    app.go                           # Main bubbletea program
    list_view.go                     # Cluster list (main view)
    detail_view.go                   # Cluster details
    create_view.go                   # Create cluster form
    styles.go                        # Lipgloss styles
  logger/
    logger.go                        # File logging
go.mod
go.sum
README.md
```

## Phase 1: Project Setup

### 1.1 Initialize Project
- [x] Create directory: `~/sideprojects/local-k8s-manager`
- [ ] Initialize git repo
- [ ] Create Go module: `go mod init github.com/[username]/local-k8s-manager`
- [ ] Install dependencies:
  ```bash
  go get github.com/charmbracelet/bubbletea
  go get github.com/charmbracelet/bubbles/table
  go get github.com/charmbracelet/bubbles/viewport
  go get github.com/charmbracelet/lipgloss
  ```

### 1.2 mise Configuration
Create `.mise.toml`:
```toml
[tools]
go = "1.23"
golangci-lint = "latest"

[tasks.test]
run = "go test ./... -v -race -cover"

[tasks.lint]
run = "golangci-lint run"

[tasks.build]
run = "go build -o bin/lkm cmd/lkm/main.go"

[tasks.run]
run = "go run cmd/lkm/main.go"

[tasks.install]
run = "go install ./cmd/lkm"
```

### 1.3 GitHub Actions
Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
      - name: go fmt check
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "Please run: go fmt ./..."
            exit 1
          fi

  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        if: matrix.go-version == '1.23'

  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Build
        run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o bin/lkm-${{ matrix.goos }}-${{ matrix.goarch }} cmd/lkm/main.go
```

### 1.4 Linter Config
Create `.golangci.yml`:
```yaml
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
```

## Phase 2: Backend Layer

### 2.1 Provider Interface
`internal/backend/interface.go`:
```go
type Provider interface {
    Name() string
    IsInstalled() bool
    List() ([]models.Cluster, error)
    Create(name string, opts CreateOptions) error
    Delete(name string) error
    Start(name string) error
    Stop(name string) error
}

type CreateOptions struct {
    Workers int
}
```

### 2.2 CLI Wrappers
Implement for each provider:

**k3d** (`internal/backend/k3d.go`):
- List: `k3d cluster list -o json`
- Create: `k3d cluster create <name> --agents <n>`
- Delete: `k3d cluster delete <name>`
- Start: `k3d cluster start <name>`
- Stop: `k3d cluster stop <name>`
- Parse JSON output

**kind** (`internal/backend/kind.go`):
- List: `kind get clusters` (plain text)
- Create: `kind create cluster --name <name>`
- Delete: `kind delete cluster --name <name>`
- Start/Stop: Not supported (show warning)
- Parse plain text output

**minikube** (`internal/backend/minikube.go`):
- List: `minikube profile list -o json`
- Create: `minikube start --profile <name>`
- Delete: `minikube delete -p <name> -o json`
- Start: `minikube start -p <name>`
- Stop: `minikube stop -p <name> -o json`
- Parse JSON output

### 2.3 Executor
`internal/backend/executor.go`:
```go
func Exec(name string, args ...string) ([]byte, error) {
    cmd := exec.Command(name, args...)
    output, err := cmd.CombinedOutput()

    // Log command execution
    logger.Log("exec", map[string]interface{}{
        "command": name,
        "args": args,
        "error": err,
    })

    return output, err
}

func IsCommandAvailable(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}
```

## Phase 3: Logging

### 3.1 Logger
`internal/logger/logger.go`:
```go
// Log to ~/.local/share/lkm/lkm.log
// Format: JSON lines
// Fields: timestamp, level, operation, details

func Init() error
func Log(operation string, details map[string]interface{})
func LogError(operation string, err error, details map[string]interface{})
```

Log operations:
- Cluster list/create/delete/start/stop
- Command execution
- Errors
- TUI state changes (optional)

## Phase 4: TUI Layer

### 4.1 Main Program
`internal/tui/app.go`:
```go
type viewState int

const (
    listView viewState = iota
    detailView
    createView
)

type model struct {
    view      viewState
    providers []backend.Provider
    clusters  []models.Cluster
    cursor    int
    width     int
    height    int
    err       error
    loading   bool
}

func (m model) Init() tea.Cmd {
    return tea.Batch(
        loadClustersCmd(m.providers),
        tickCmd(),
    )
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyPress(msg)
    case clustersLoadedMsg:
        m.clusters = msg.clusters
        m.loading = false
    case tickMsg:
        return m, tea.Batch(
            loadClustersCmd(m.providers),
            tickCmd(),
        )
    }
    return m, nil
}

func (m model) View() string {
    switch m.view {
    case listView:
        return renderListView(m)
    case detailView:
        return renderDetailView(m)
    case createView:
        return renderCreateView(m)
    }
    return ""
}
```

### 4.2 List View
`internal/tui/list_view.go`:
- Table with columns: Provider | Name | Status | Nodes | Age
- Keybindings:
  - `j/k` or `↑/↓`: Navigate
  - `c`: Create cluster
  - `d`: Delete selected
  - `s`: Start selected
  - `x`: Stop selected
  - `Enter`: Detail view
  - `r`: Refresh now
  - `?`: Help
  - `q`: Quit
- Auto-refresh every 5s
- Loading spinner during refresh

### 4.3 Detail View
`internal/tui/detail_view.go`:
- Show: Name, Provider, Status, Nodes, API endpoint, Kubeconfig
- Keybindings:
  - `Esc`: Back to list
  - `d`: Delete cluster
  - `s`: Start cluster
  - `x`: Stop cluster

### 4.4 Create View
`internal/tui/create_view.go`:
- Form:
  - Provider selection (radio buttons)
  - Cluster name (text input)
  - Worker nodes (number input, default 1)
- Keybindings:
  - `Tab`: Next field
  - `Enter`: Submit
  - `Esc`: Cancel

### 4.5 Styles
`internal/tui/styles.go`:
- Use lipgloss for consistent theming
- Colors: green (running), yellow (stopped), red (error), gray (unknown)
- Borders, padding, margins

## Phase 5: Models

### 5.1 Cluster Model
`internal/models/cluster.go`:
```go
type Cluster struct {
    Name      string
    Provider  string
    Status    Status
    Nodes     int
    CreatedAt time.Time
}

type Status string

const (
    StatusRunning Status = "running"
    StatusStopped Status = "stopped"
    StatusUnknown Status = "unknown"
    StatusError   Status = "error"
)
```

## Phase 6: Entry Point

### 6.1 Main
`cmd/lkm/main.go`:
```go
func main() {
    // Initialize logger
    if err := logger.Init(); err != nil {
        fmt.Printf("Warning: logging disabled: %v\n", err)
    }

    providers := []backend.Provider{
        backend.NewK3dProvider(),
        backend.NewKindProvider(),
        backend.NewMinikubeProvider(),
    }

    // Filter to installed providers
    var available []backend.Provider
    for _, p := range providers {
        if p.IsInstalled() {
            available = append(available, p)
        }
    }

    if len(available) == 0 {
        fmt.Println("No cluster tools found.")
        fmt.Println("Install one of: k3d, kind, minikube")
        os.Exit(1)
    }

    logger.Log("startup", map[string]interface{}{
        "providers": getProviderNames(available),
    })

    p := tea.NewProgram(
        tui.NewModel(available),
        tea.WithAltScreen(),
    )

    if _, err := p.Run(); err != nil {
        logger.LogError("fatal", err, nil)
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Implementation Phases

### Phase 1: Project Setup + Logging (Week 1)
1. Project setup (mise, GitHub Actions, golangci)
2. Logger implementation
3. Go module and dependencies

### Phase 2: Backend (Week 1-2)
4. Backend interface
5. k3d implementation
6. Executor with logging

### Phase 3: Basic TUI (Week 2)
7. Basic TUI structure
8. List view (manual refresh only)
9. Navigation with j/k

### Phase 4: Operations (Week 2-3)
10. Delete operation with confirmation
11. Start/Stop operations
12. Error handling and display
13. Auto-refresh every 5s

### Phase 5: Create + Other Providers (Week 3)
14. Create view with form
15. kind provider implementation
16. minikube provider implementation
17. Provider auto-detection

### Phase 6: Detail View (Week 4)
18. Detail view layout
19. Cluster info fetching
20. Navigation between views

### Phase 7: Polish (Week 4-5)
21. Loading spinners
22. Status colors and styling
23. Help menu (?)
24. README and documentation
25. Testing

## Testing Strategy

### Unit Tests
- Backend: Mock `exec.Command` via interface
- Test each provider's command building
- Test JSON/text parsing
- Logger tests

### Integration Tests
- Require k3d/kind/minikube installed
- Create test clusters
- Verify operations work
- Clean up after tests
- Check logs written correctly

### Manual Testing
- Test with each provider
- Test with multiple clusters
- Test error scenarios (tool not found, cluster not found)
- Test UI responsiveness
- Verify logging works

## Distribution

### Local Build
```bash
mise run build
./bin/lkm
```

### Install
```bash
mise run install
lkm
```

### goreleaser (Optional - Future)
Create `.goreleaser.yml` for:
- Multi-platform builds
- GitHub releases
- Homebrew tap
- Archive creation

## Logging Details

### Log Location
`~/.local/share/lkm/lkm.log` (or `$XDG_DATA_HOME/lkm/lkm.log`)

### Log Format
JSON lines:
```json
{"timestamp":"2025-12-29T10:00:00Z","level":"info","operation":"cluster.create","details":{"provider":"k3d","name":"test","workers":2}}
{"timestamp":"2025-12-29T10:00:05Z","level":"error","operation":"cluster.delete","error":"cluster not found","details":{"provider":"kind","name":"foo"}}
```

### Log Rotation
- Max size: 10MB
- Keep last 3 files
- Auto-rotate on size limit

### Logged Operations
- startup (providers available)
- cluster.list
- cluster.create (name, provider, options)
- cluster.delete (name, provider)
- cluster.start (name, provider)
- cluster.stop (name, provider)
- exec (command, args, duration, error)
- fatal (error)

## References

- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [k3d Docs](https://k3d.io/v5.3.0/)
- [kind Docs](https://kind.sigs.k8s.io/)
- [minikube Docs](https://minikube.sigs.k8s.io/docs/)
- [lazydocker](https://github.com/jesseduffield/lazydocker) - Reference implementation

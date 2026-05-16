package pkg

type MockManager struct {
	AvailableTrue bool
	InstalledMap  map[string]bool
	InstallCalls  [][]string
	InstallErr    error
}

func (m *MockManager) Name() string { return "mock" }

func (m *MockManager) IsAvailable() bool { return m.AvailableTrue }

func (m *MockManager) IsInstalled(pkg string) bool {
	if m.InstalledMap == nil {
		return false
	}
	return m.InstalledMap[pkg]
}

func (m *MockManager) Install(pkgs ...string) error {
	m.InstallCalls = append(m.InstallCalls, pkgs)
	return m.InstallErr
}

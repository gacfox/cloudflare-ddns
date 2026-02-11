package status

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type IPStatus struct {
	LastIP            string   `json:"last_ip"`
	LastUpdateDomains []string `json:"last_update_domains"`
}

type Status struct {
	IPv4 IPStatus `json:"ipv4"`
	IPv6 IPStatus `json:"ipv6"`
}

type Manager struct {
	statusFilePath string
}

func NewManager() *Manager {
	tempDir := os.TempDir()
	statusDir := filepath.Join(tempDir, "cloudflare-ddns")
	statusFilePath := filepath.Join(statusDir, "status.dat")

	return &Manager{
		statusFilePath: statusFilePath,
	}
}

func (m *Manager) Load() (*Status, error) {
	status := newEmptyStatus()

	if _, err := os.Stat(m.statusFilePath); os.IsNotExist(err) {
		return status, nil
	}

	data, err := os.ReadFile(m.statusFilePath)
	if err != nil {
		logrus.Errorf("读取临时文件status.dat错误: %v", err)
		return status, nil
	}

	if err := json.Unmarshal(data, status); err != nil {
		logrus.Errorf("解析临时文件status.dat错误: %v", err)
		return status, nil
	}

	return status, nil
}

func (m *Manager) InitializeEmptyStatus() *Status {
	return newEmptyStatus()
}

func GetIPStatus(st *Status, ipType string) *IPStatus {
	switch ipType {
	case "ipv4":
		return &st.IPv4
	case "ipv6":
		return &st.IPv6
	}
	return nil
}

func (m *Manager) Save(status *Status) error {
	statusDir := filepath.Dir(m.statusFilePath)
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		logrus.Errorf("创建状态目录错误: %v", err)
		return err
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		logrus.Errorf("序列化状态文件错误: %v", err)
		return err
	}

	if err := os.WriteFile(m.statusFilePath, data, 0644); err != nil {
		logrus.Errorf("写入临时文件status.dat错误: %v", err)
		return err
	}

	return nil
}

func newEmptyStatus() *Status {
	return &Status{
		IPv4: IPStatus{
			LastIP:            "",
			LastUpdateDomains: []string{},
		},
		IPv6: IPStatus{
			LastIP:            "",
			LastUpdateDomains: []string{},
		},
	}
}

package main

import (
	"cloudflare-ddns/internal/api"
	"cloudflare-ddns/internal/config"
	"cloudflare-ddns/internal/network"
	"cloudflare-ddns/internal/status"
	"cloudflare-ddns/internal/updater"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := loadAndValidateConfig()
	if err != nil {
		logrus.Fatalf("加载配置失败: %v", err)
	}

	setupLogger(cfg)

	logrus.Infof("配置加载成功，ZoneID: %s, 网络接口: %s, 域名: %v",
		cfg.ZoneID, cfg.NetworkInterface, cfg.DomainNames)

	client := api.NewClient(cfg.ZoneID, cfg.AuthorizationKey)
	statusManager := status.NewManager()
	if err := runUpdate(cfg, statusManager, client); err != nil {
		logrus.Fatalf("首次更新失败: %v", err)
	}

	logrus.Infof("定时任务启动，更新间隔: %d 秒", cfg.UpdateIntervalSeconds)

	for {
		time.Sleep(time.Duration(cfg.UpdateIntervalSeconds) * time.Second)

		latestCfg, err := loadAndValidateConfig()
		if err != nil {
			logrus.Errorf("热加载配置失败，继续使用旧配置: %v", err)
		} else {
			cfg = latestCfg
			setupLogger(cfg)
			client = api.NewClient(cfg.ZoneID, cfg.AuthorizationKey)
			logrus.Infof("配置热加载成功，ZoneID: %s, 网络接口: %s, 域名: %v, 更新间隔: %d 秒",
				cfg.ZoneID, cfg.NetworkInterface, cfg.DomainNames, cfg.UpdateIntervalSeconds)
		}

		if err := runUpdate(cfg, statusManager, client); err != nil {
			logrus.Errorf("更新失败: %v", err)
		}
	}
}

func runUpdate(cfg *config.Config, statusManager *status.Manager, client *api.Client) error {
	upd := updater.NewUpdater(cfg, client)

	ipv4, ipv6, err := network.GetNetworkAddresses(cfg.NetworkInterface)
	if err != nil {
		logrus.Errorf("获取网络接口[%s]地址失败: %v", cfg.NetworkInterface, err)
		return err
	}

	logrus.Infof("读取网络接口[%s]: 公网IPv4 [%s] 公网IPv6 [%s]", cfg.NetworkInterface, ipv4, ipv6)

	currentStatus, err := statusManager.Load()
	if err != nil {
		logrus.Errorf("读取状态文件失败: %v", err)
		currentStatus = statusManager.InitializeEmptyStatus()
	}

	var updatedDomainsIPv4 []string
	var updatedDomainsIPv6 []string

	if cfg.IPv4DDNS {
		if ipv4 != "" {
			updatedDomainsIPv4 = upd.ProcessUpdate(status.GetIPStatus(currentStatus, "ipv4"), "A", ipv4)
		} else {
			logrus.Warnf("无法获取网络接口[%s]IPv4地址", cfg.NetworkInterface)
		}
	}

	if cfg.IPv6DDNS {
		if ipv6 != "" {
			updatedDomainsIPv6 = upd.ProcessUpdate(status.GetIPStatus(currentStatus, "ipv6"), "AAAA", ipv6)
		} else {
			logrus.Warnf("无法获取网络接口[%s]IPv6地址", cfg.NetworkInterface)
		}
	}

	newStatus := &status.Status{
		IPv4: status.IPStatus{
			LastIP:            ipv4,
			LastUpdateDomains: updatedDomainsIPv4,
		},
		IPv6: status.IPStatus{
			LastIP:            ipv6,
			LastUpdateDomains: updatedDomainsIPv6,
		},
	}

	if err := statusManager.Save(newStatus); err != nil {
		logrus.Errorf("保存状态文件失败: %v", err)
	}

	return nil
}

func setupLogger(cfg *config.Config) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
	})

	level, err := logrus.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
}

func loadAndValidateConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

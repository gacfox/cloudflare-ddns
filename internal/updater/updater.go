package updater

import (
	"cloudflare-ddns/internal/api"
	"cloudflare-ddns/internal/config"
	"cloudflare-ddns/internal/status"
	"slices"

	"github.com/sirupsen/logrus"
)

type Updater struct {
	config *config.Config
	client *api.Client
}

func NewUpdater(cfg *config.Config, client *api.Client) *Updater {
	return &Updater{
		config: cfg,
		client: client,
	}
}

func (u *Updater) ProcessUpdate(status *status.IPStatus, recordType string, ip string) []string {
	updatedDomains := slices.Clone(status.LastUpdateDomains)

	toUpdateDomains := u.determineDomainsToUpdate(status, ip)

	if len(toUpdateDomains) == 0 {
		logrus.Infof("没有需要更新的域名")
		return updatedDomains
	}

	for _, domain := range toUpdateDomains {
		err := u.updateDomain(domain, recordType, ip)
		if err != nil {
			logrus.Errorf("更新DNS记录异常: %v", err)
			continue
		}

		if !contains(updatedDomains, domain) {
			updatedDomains = append(updatedDomains, domain)
		}
	}

	return updatedDomains
}

func (u *Updater) determineDomainsToUpdate(status *status.IPStatus, ip string) []string {
	var toUpdateDomains []string

	if status.LastIP == ip {
		for _, domain := range u.config.DomainNames {
			if !contains(status.LastUpdateDomains, domain) {
				toUpdateDomains = append(toUpdateDomains, domain)
			}
		}
	} else {
		toUpdateDomains = append(toUpdateDomains, u.config.DomainNames...)
	}

	return toUpdateDomains
}

func (u *Updater) updateDomain(domain string, recordType string, ip string) error {
	dnsRecords, err := u.client.GetDNSRecord(recordType, domain)
	if err != nil {
		return err
	}

	if len(dnsRecords) == 0 {
		logrus.Infof("域名[%s]不存在,创建新的DNS记录 -> [%s]", domain, ip)
		_, err = u.client.CreateDNSRecord(recordType, domain, ip)
		return err
	}

	dnsRecordID := dnsRecords[0].ID
	logrus.Infof("域名[%s]存在,更新DNS记录 -> [%s]", domain, ip)
	return u.client.UpdateDNSRecord(recordType, domain, ip, dnsRecordID)
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

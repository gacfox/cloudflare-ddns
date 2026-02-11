package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	baseURL    = "https://api.cloudflare.com/client/v4"
	timeoutSec = 30
	ttl        = 120
)

type Client struct {
	ZoneID           string
	AuthorizationKey string
	httpClient       *http.Client
}

func NewClient(zoneID, authorizationKey string) *Client {
	return &Client{
		ZoneID:           zoneID,
		AuthorizationKey: authorizationKey,
		httpClient: &http.Client{
			Timeout: timeoutSec * time.Second,
		},
	}
}

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
}

type APIResponseList struct {
	Success bool        `json:"success"`
	Errors  []APIError  `json:"errors"`
	Result  []DNSRecord `json:"result"`
}

type APIResponseSingle struct {
	Success bool       `json:"success"`
	Errors  []APIError `json:"errors"`
	Result  DNSRecord  `json:"result"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) GetDNSRecord(recordType, domain string) ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records?type=%s&name=%s&match=all",
		baseURL, c.ZoneID, recordType, domain)

	logrus.Debugf("Cloudflare API GET dns_records: type=%s domain=%s url=%s", recordType, domain, url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthorizationKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponseList
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		code, message := getErrorDescription(apiResp.Errors)
		return nil, fmt.Errorf("API调用失败：错误码：%d 错误描述：%s", code, message)
	}

	logrus.Debugf("Cloudflare API GET dns_records success: type=%s domain=%s count=%d", recordType, domain, len(apiResp.Result))

	if len(apiResp.Result) == 0 {
		return []DNSRecord{}, nil
	}

	return apiResp.Result, nil
}

func (c *Client) CreateDNSRecord(recordType, domain, ip string) (*DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records", baseURL, c.ZoneID)

	record := DNSRecord{
		Type:    recordType,
		Name:    domain,
		Content: ip,
		TTL:     ttl,
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	logrus.Debugf("Cloudflare API POST dns_records: type=%s domain=%s ip=%s url=%s", recordType, domain, ip, url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthorizationKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponseSingle
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		code, message := getErrorDescription(apiResp.Errors)
		return nil, fmt.Errorf("API调用失败：错误码：%d 错误描述：%s", code, message)
	}

	if apiResp.Result.ID != "" {
		logrus.Debugf("Cloudflare API POST dns_records success: type=%s domain=%s id=%s", recordType, domain, apiResp.Result.ID)
		return &apiResp.Result, nil
	}

	return nil, fmt.Errorf("创建 DNS 记录失败：返回结果为空")
}

func (c *Client) UpdateDNSRecord(recordType, domain, ip, dnsID string) error {
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", baseURL, c.ZoneID, dnsID)

	record := DNSRecord{
		Type:    recordType,
		Name:    domain,
		Content: ip,
		TTL:     ttl,
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return err
	}

	logrus.Debugf("Cloudflare API PUT dns_records: type=%s domain=%s ip=%s id=%s url=%s", recordType, domain, ip, dnsID, url)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.AuthorizationKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp APIResponseSingle
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return err
	}

	if !apiResp.Success {
		code, message := getErrorDescription(apiResp.Errors)
		return fmt.Errorf("API调用失败：错误码：%d 错误描述：%s", code, message)
	}

	logrus.Debugf("Cloudflare API PUT dns_records success: type=%s domain=%s id=%s", recordType, domain, dnsID)
	return nil
}

func getErrorDescription(errors []APIError) (int, string) {
	if len(errors) > 0 {
		return errors[0].Code, errors[0].Message
	}
	return 0, "None"
}

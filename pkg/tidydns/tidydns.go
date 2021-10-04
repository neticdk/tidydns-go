package tidydns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TidyDNSClient interface {
	ListZones(ctx context.Context) ([]*ZoneInfo, error)
	FindZoneID(ctx context.Context, name string) (int, error)
	CreateRecord(ctx context.Context, zoneID int, info RecordInfo) (int, error)
	UpdateRecord(ctx context.Context, zoneID int, recordID int, info RecordInfo) error
	ReadRecord(ctx context.Context, zoneID int, recordID int) (*RecordInfo, error)
	FindRecord(ctx context.Context, zoneID int, name string, rType RecordType) ([]*RecordInfo, error)
	ListRecords(ctx context.Context, zoneID int) ([]*RecordInfo, error)
	DeleteRecord(ctx context.Context, zoneID int, recordID int) error
}

type ZoneInfo struct {
	ID   int
	Name string
}

type SubnetIDs struct {
	SubnetID int
	ZoneID   int
	VlanNo   int
}

type InterfaceInfo struct {
	InterfaceIP   string
	Interfacename string
}

type CreateInfo struct {
	SubnetID      int
	ZoneID        int
	InterfaceIP   string
	InterfaceName string
}

type RecordInfo struct {
	ID          int
	Type        RecordType
	Name        string
	Description string
	Destination string
	TTL         int
	Status      RecordStatus
	Location    LocationID
}

type LocationID int
type RecordType int
type RecordStatus int

const (
	RecordStatusActive   RecordStatus = 0
	RecordStatusInactive RecordStatus = 1
	RecordStatusDeleted  RecordStatus = 2

	RecordTypeA     RecordType = 0
	RecordTypeAPTR  RecordType = 1
	RecordTypeCNAME RecordType = 2
	RecordTypeMX    RecordType = 3
	RecordTypeNS    RecordType = 4
	RecordTypeTXT   RecordType = 5
	RecordTypeSRV   RecordType = 6
	RecordTypeDS    RecordType = 7
	RecordTypeSSHFP RecordType = 8
	RecordTypeTLSA  RecordType = 9
	RecordTypeCAA   RecordType = 10
)

type tidyDNSClient struct {
	client   *http.Client
	username string
	password string
	baseURL  string
}

func New(baseURL, username, password string) TidyDNSClient {
	return &tidyDNSClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

func (c *tidyDNSClient) ListZones(ctx context.Context) ([]*ZoneInfo, error) {
	var zones []zoneInfo
	err := c.getData(ctx, fmt.Sprintf("%s/=/zone?type=json", c.baseURL), &zones)
	if err != nil {
		return nil, err
	}

	result := make([]*ZoneInfo, 0)
	for _, zone := range zones {
		result = append(result, &ZoneInfo{
			ID:   zone.ID,
			Name: zone.Name,
		})
	}
	return result, nil
}

func (c *tidyDNSClient) FindZoneID(ctx context.Context, name string) (int, error) {
	var zones []zoneInfo
	err := c.getData(ctx, fmt.Sprintf("%s/=/zone?type=json&name=%s", c.baseURL, name), &zones)
	if err != nil {
		return 0, err
	}

	if len(zones) == 0 {
		return 0, fmt.Errorf("zone not found for: %s", name)
	}

	for _, z := range zones {
		if z.Name == name {
			return z.ID, nil
		}
	}

	return 0, fmt.Errorf("unable to match zone name: %s", name)
}

func (c *tidyDNSClient) CreateRecord(ctx context.Context, zoneID int, info RecordInfo) (int, error) {
	data := url.Values{
		"type":        {strconv.Itoa(int(info.Type))},
		"name":        {info.Name},
		"ttl":         {strconv.Itoa(info.TTL)},
		"description": {info.Description},
		"status":      {strconv.Itoa(int(info.Status))},
		"destination": {info.Destination},
		"location_id": {strconv.Itoa(int(info.Location))},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/=/record/new/%d", c.baseURL, zoneID), strings.NewReader(data.Encode()))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/=/record_merged?zone_id=%d", c.baseURL, zoneID), nil)
	if err != nil {
		return 0, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err = c.client.Do(req)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	var records []recordList
	err = json.NewDecoder(res.Body).Decode(&records)
	if err != nil {
		return 0, err
	}

	for _, r := range records {
		if r.Type == info.Type && r.Name == info.Name && r.Destination == info.Destination {
			return r.ID, nil
		}
	}

	return 0, fmt.Errorf("unable to find new record")
}

func (c *tidyDNSClient) UpdateRecord(ctx context.Context, zoneID int, recordID int, info RecordInfo) error {
	data := url.Values{
		"ttl":         {strconv.Itoa(info.TTL)},
		"description": {info.Description},
		"status":      {strconv.Itoa(int(info.Status))},
		"destination": {info.Destination},
		"location_id": {strconv.Itoa(int(info.Location))},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, recordID, zoneID), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	return nil
}

func (c *tidyDNSClient) FindRecord(ctx context.Context, zoneID int, name string, rType RecordType) ([]*RecordInfo, error) {
	var records []recordList
	err := c.getData(ctx, fmt.Sprintf("%s/=/record?type=json&zone=%d&name=%s", c.baseURL, zoneID, name), &records)
	if err != nil {
		return nil, err
	}

	result := make([]*RecordInfo, 0)
	for _, r := range records {
		if r.Type == rType {
			result = append(result, &RecordInfo{
				ID:          r.ID,
				Type:        r.Type,
				Name:        r.Name,
				Description: r.Description,
				Destination: r.Destination,
				TTL:         r.TTL,
				Location:    r.Location,
			})
		}
	}
	return result, nil
}

func (c *tidyDNSClient) ListRecords(ctx context.Context, zoneID int) ([]*RecordInfo, error) {
	var records []recordList
	err := c.getData(ctx, fmt.Sprintf("%s/=/record_merged?type=json&zone_id=%d&showall=1", c.baseURL, zoneID), &records)
	if err != nil {
		return nil, err
	}

	result := make([]*RecordInfo, 0)
	for _, r := range records {
		result = append(result, &RecordInfo{
			ID:          r.ID,
			Type:        r.Type,
			Name:        r.Name,
			Description: r.Description,
			Destination: r.Destination,
			TTL:         r.TTL,
			Location:    r.Location,
		})
	}
	return result, nil
}

func (c *tidyDNSClient) ReadRecord(ctx context.Context, zoneID int, recordID int) (*RecordInfo, error) {
	var record recordRead
	err := c.getData(ctx, fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, zoneID, recordID), &record)
	if err != nil {
		return nil, err
	}

	return &RecordInfo{
		ID:          record.ID,
		Type:        record.Type,
		Name:        record.Name,
		Description: record.Description,
		Destination: record.Destination,
		TTL:         record.TTL,
		Status:      record.Status,
		Location:    record.Location,
	}, nil
}

func (c *tidyDNSClient) DeleteRecord(ctx context.Context, zoneID int, recordID int) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, recordID, zoneID), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	return nil
}

func (c *tidyDNSClient) getData(ctx context.Context, url string, value interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

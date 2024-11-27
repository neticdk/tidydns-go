package tidydns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TidyDNSClient interface {
	GetSubnetIDs(ctx context.Context, subnetCIDR string) (*SubnetIDs, error)
	GetFreeIP(ctx context.Context, subnetID int) (string, error)
	CreateDHCPInterface(ctx context.Context, createInfo CreateInfo) (int, error)
	ReadDHCPInterface(ctx context.Context, interfaceID int) (*InterfaceInfo, error)
	UpdateDHCPInterfaceName(ctx context.Context, interfaceID int, interfaceName string) (int, error)
	DeleteDHCPInterface(ctx context.Context, interfaceID int) error
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
	LocationID    int
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

//goland:noinspection GoUnusedConst
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

func (c *tidyDNSClient) GetSubnetIDs(ctx context.Context, subnetCIDR string) (*SubnetIDs, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/=/dhcp_subnet?subnet=%s", c.baseURL, subnetCIDR), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	// Check response code

	var subnets []dhcpSubnet
	err = json.NewDecoder(res.Body).Decode(&subnets)
	if err != nil {
		return nil, err
	}

	if len(subnets) == 0 {
		return nil, fmt.Errorf("subnet not found: %s", subnetCIDR)
	}

	if len(subnets) > 1 {
		return nil, fmt.Errorf("too many subnets found: %s", subnetCIDR)
	}

	return &SubnetIDs{
		SubnetID: subnets[0].ID,
		ZoneID:   subnets[0].ZoneID,
		VlanNo:   subnets[0].VlanNo,
	}, nil
}

func (c *tidyDNSClient) GetFreeIP(ctx context.Context, subnetID int) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/=/dhcp_subnet_free_ip/%d", c.baseURL, subnetID), nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	var freeIP dhcpFreeIP
	err = json.NewDecoder(res.Body).Decode(&freeIP)
	if err != nil {
		return "", err
	}

	return freeIP.Data.IPAddress, nil
}

func (c *tidyDNSClient) CreateDHCPInterface(ctx context.Context, createInfo CreateInfo) (int, error) {
	data := url.Values{
		"subnet_id":   {strconv.Itoa(createInfo.SubnetID)},
		"zone_id":     {strconv.Itoa(createInfo.ZoneID)},
		"name":        {createInfo.InterfaceName},
		"destination": {createInfo.InterfaceIP},
		"location_id": {strconv.Itoa(createInfo.LocationID)},
	}

	var checkstring string = fmt.Sprintf("Key (destination)=(%s) already exists", createInfo.InterfaceIP)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/=/dhcp_interface//new", c.baseURL), strings.NewReader(data.Encode()))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, _ := c.client.Do(req)

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		if err != nil {
			return 0, err
		}
		if strings.Contains(bodyString, checkstring) {
			return 1, fmt.Errorf("try again")
		}
		return 0, fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	var createResp interfaceCreate
	err = json.NewDecoder(res.Body).Decode(&createResp)
	if err != nil {
		return 0, err
	}

	return createResp.ID, nil
}

func (c *tidyDNSClient) ReadDHCPInterface(ctx context.Context, interfaceID int) (*InterfaceInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/=/dhcp_interface/?id=%d", c.baseURL, interfaceID), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from tidyDNS server: %s", res.Status)
	}

	var interfaceRead interfaceRead
	err = json.NewDecoder(res.Body).Decode(&interfaceRead)
	if err != nil {
		return nil, err
	}

	return &InterfaceInfo{
		InterfaceIP:   interfaceRead.Destination,
		Interfacename: interfaceRead.Name,
	}, nil
}

func (c *tidyDNSClient) UpdateDHCPInterfaceName(ctx context.Context, interfaceID int, interfaceName string) (int, error) {
	data := url.Values{
		"name": {interfaceName},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/=/dhcp_interface//%d", c.baseURL, interfaceID), strings.NewReader(data.Encode()))
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

	var createResp interfaceCreate
	err = json.NewDecoder(res.Body).Decode(&createResp)
	if err != nil {
		return 0, err
	}

	return createResp.ID, nil
}

func (c *tidyDNSClient) DeleteDHCPInterface(ctx context.Context, interfaceID int) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/=/dhcp_interface/%d", c.baseURL, interfaceID), nil)
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
	if res != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if err != nil || res == nil {
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
	if res != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if err != nil || res == nil {
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
	if res != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if err != nil || res == nil {
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
		if r.Type == rType && r.Name == name {
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
	if res != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if err != nil || res == nil {
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
	if res != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
	}

	if err != nil || res == nil {
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

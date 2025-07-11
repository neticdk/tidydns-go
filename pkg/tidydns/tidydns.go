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
	"time"
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
	CreateInternalUser(ctx context.Context, username string, password string, description string, changePasswordOnFirstLogin bool, authGroup AuthGroup, userAllow []UserAllowID) (UserID, error)
	GetInternalUser(ctx context.Context, userID UserID) (*UserInfo, error)
	UpdateInternalUser(ctx context.Context, userID UserID, password *string, description *string, authGroup *AuthGroup, userAllow []UserAllowID) error
	DeleteInternalUser(ctx context.Context, userID UserID) error
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

type UserInfo struct {
	ModifiedBy        string
	Description       string
	ModifiedDate      time.Time
	Username          string
	AuthGroup         AuthGroup
	Name              string
	PasswdChangedDate time.Time
	Id                UserID
	Groups            []UserInfoGroup
}

type UserInfoGroup struct {
	GroupName   string  `json:"groupname"`
	Name        string  `json:"name"`
	Notes       *string `json:"notes,omitempty"`
	Id          int     `json:"id"`
	Description *string `json:"description,omitempty"`
}

type UserID int
type LocationID int
type RecordType int
type RecordStatus int
type AuthGroup int
type UserAllowID int

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

	AuthGroupUser       AuthGroup = 2
	AuthGroupSuperAdmin AuthGroup = 1
)

const errorTidyDNS = "error from tidyDNS server: %s"
const headerContentType = "Content-Type"
const mimeForm = "application/x-www-form-urlencoded"

type tidyDNSClient struct {
	client   *http.Client
	username string
	password string
	baseURL  string
}

func (c *tidyDNSClient) CreateInternalUser(ctx context.Context, username string, password string, description string, changePasswordOnFirstLogin bool, authGroup AuthGroup, userAllow []UserAllowID) (UserID, error) {
	var userAllowFormatted []string
	if len(userAllow) > 0 {
		userAllowFormatted = make([]string, 0, len(userAllowFormatted))
		for _, id := range userAllow {
			userAllowFormatted = append(userAllowFormatted, strconv.Itoa(int(id)))
		}
	} else {
		userAllowFormatted = []string{""}
	}

	var changePasswordOnFirstLoginFormatted string
	if changePasswordOnFirstLogin {
		changePasswordOnFirstLoginFormatted = "1"
	} else {
		changePasswordOnFirstLoginFormatted = "0"
	}

	data := url.Values{
		"username":                       {username},
		"epassword":                      {password},
		"epassword_verify":               {password},
		"change_password_on_first_login": {changePasswordOnFirstLoginFormatted},
		"description":                    {description},
		"auth_group":                     {strconv.Itoa(int(authGroup))},
		//"tmp_auth_group":                 {""},
		"user_allow": userAllowFormatted,
	}

	newUserUrl := fmt.Sprintf("%s/=/user/new", c.baseURL)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		newUserUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		if err != nil {
			return 0, err
		}
		if strings.Contains(bodyString, fmt.Sprintf("Key (username)=(%s) already exists", username)) {
			return 0, fmt.Errorf("username already exists")
		}
		return 0, fmt.Errorf(errorTidyDNS, res.Status)
	}

	var user userCreate
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return 0, err
	}

	return UserID(user.Data.Id), nil
}

func (c *tidyDNSClient) GetInternalUser(ctx context.Context, userID UserID) (*UserInfo, error) {
	userLookupUrl := fmt.Sprintf("%s/=/user/%s", c.baseURL, strconv.Itoa(int(userID)))
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		userLookupUrl,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errorTidyDNS, res.Status)
	}

	var user userRead
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	modifiedDate, err := time.Parse(time.DateTime, user.ModifiedDate)
	if err != nil {
		return nil, err
	}

	passwordChangedDate, err := time.Parse(time.DateTime, user.PasswdChangedDate)
	if err != nil {
		return nil, err
	}

	var ag AuthGroup
	switch user.AuthGroup {
	case "User":
		ag = AuthGroupUser
	case "SuperAdmin":
		ag = AuthGroupSuperAdmin
	default:
		return nil, fmt.Errorf("unknown auth group")
	}

	return &UserInfo{
		ModifiedBy:        user.ModifiedBy,
		Description:       user.Description,
		ModifiedDate:      modifiedDate,
		Username:          user.Username,
		AuthGroup:         ag,
		Name:              user.Name,
		PasswdChangedDate: passwordChangedDate,
		Id:                UserID(user.Id),
		Groups:            user.Groups,
	}, nil
}

func (c *tidyDNSClient) UpdateInternalUser(ctx context.Context, userID UserID, password *string, description *string, authGroup *AuthGroup, userAllow []UserAllowID) error {
	data := url.Values{}

	if password != nil {
		data.Set("epassword", *password)
		data.Set("epassword_verify", *password)
	}

	if description != nil {
		data.Set("description", *description)
	}

	if authGroup != nil {
		data.Set("auth_group", strconv.Itoa(int(*authGroup)))
	}

	if userAllow != nil {
		var userAllowFormatted []string
		if len(userAllow) > 0 {
			userAllowFormatted = make([]string, 0, len(userAllowFormatted))
			for _, id := range userAllow {
				userAllowFormatted = append(userAllowFormatted, strconv.Itoa(int(id)))
			}
		} else {
			userAllowFormatted = []string{""}
		}
		data["user_allow"] = userAllowFormatted
	}

	userLookupUrl := fmt.Sprintf("%s/=/user/%s", c.baseURL, strconv.Itoa(int(userID)))
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		userLookupUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
	}

	var user userCreate
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return err
	}

	return nil
}

func (c *tidyDNSClient) DeleteInternalUser(ctx context.Context, userID UserID) error {
	userLookupUrl := fmt.Sprintf("%s/=/user/%s", c.baseURL, strconv.Itoa(int(userID)))
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		userLookupUrl,
		nil,
	)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
	}

	return nil
}

func New(baseURL, username, password string) TidyDNSClient {
	return &tidyDNSClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   &http.Client{},
	}
}

func closeResponse(resp *http.Response) {
	if resp != nil {
		_ = resp.Body.Close()
	}
}

func (c *tidyDNSClient) GetSubnetIDs(ctx context.Context, subnetCIDR string) (*SubnetIDs, error) {
	dhcpSubnetUrl := fmt.Sprintf("%s/=/dhcp_subnet?subnet=%s", c.baseURL, subnetCIDR)
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		dhcpSubnetUrl,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errorTidyDNS, res.Status)
	}

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
	dhcpFreeIPUrl := fmt.Sprintf("%s/=/dhcp_subnet_free_ip/%d", c.baseURL, subnetID)
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		dhcpFreeIPUrl,
		nil,
	)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf(errorTidyDNS, res.Status)
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

	var checkstring = fmt.Sprintf("Key (destination)=(%s) already exists", createInfo.InterfaceIP)

	dhcpInterfaceNewUrl := fmt.Sprintf("%s/=/dhcp_interface//new", c.baseURL)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		dhcpInterfaceNewUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer closeResponse(res)

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		bodyString := string(bodyBytes)
		if err != nil {
			return 0, err
		}
		if strings.Contains(bodyString, checkstring) {
			return 1, fmt.Errorf("try again")
		}
		return 0, fmt.Errorf(errorTidyDNS, res.Status)
	}

	var createResp interfaceCreate
	err = json.NewDecoder(res.Body).Decode(&createResp)
	if err != nil {
		return 0, err
	}

	return createResp.ID, nil
}

func (c *tidyDNSClient) ReadDHCPInterface(ctx context.Context, interfaceID int) (*InterfaceInfo, error) {
	dhcpInterfaceReadUrl := fmt.Sprintf("%s/=/dhcp_interface/?id=%d", c.baseURL, interfaceID)
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		dhcpInterfaceReadUrl,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(errorTidyDNS, res.Status)
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

	dhcpInterfaceLookupUrl := fmt.Sprintf("%s/=/dhcp_interface//%d", c.baseURL, interfaceID)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		dhcpInterfaceLookupUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(errorTidyDNS, res.Status)
	}

	var createResp interfaceCreate
	err = json.NewDecoder(res.Body).Decode(&createResp)
	if err != nil {
		return 0, err
	}

	return createResp.ID, nil
}

func (c *tidyDNSClient) DeleteDHCPInterface(ctx context.Context, interfaceID int) error {
	dhcpInterfaceLookupUrl := fmt.Sprintf("%s/=/dhcp_interface/%d", c.baseURL, interfaceID)
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		dhcpInterfaceLookupUrl,
		nil,
	)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
	}

	return nil
}

func (c *tidyDNSClient) ListZones(ctx context.Context) ([]*ZoneInfo, error) {
	var zones []zoneInfo
	zoneListUrl := fmt.Sprintf("%s/=/zone?type=json", c.baseURL)
	err := c.getData(
		ctx,
		zoneListUrl,
		&zones,
	)
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
	zoneLookupUrl := fmt.Sprintf("%s/=/zone?type=json&name=%s", c.baseURL, name)
	err := c.getData(
		ctx,
		zoneLookupUrl,
		&zones,
	)
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

	newRecordUrl := fmt.Sprintf("%s/=/record/new/%d", c.baseURL, zoneID)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		newRecordUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil || res == nil {
		return 0, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(errorTidyDNS, res.Status)
	}

	recordMergeUrl := fmt.Sprintf("%s/=/record_merged?type=json&zone_id=%d&showall=1", c.baseURL, zoneID)
	req, err = http.NewRequestWithContext(
		ctx,
		"GET",
		recordMergeUrl,
		nil,
	)
	if err != nil {
		return 0, err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err = c.client.Do(req)
	if err != nil || res == nil {
		return 0, err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(errorTidyDNS, res.Status)
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

	zoneLookupUrl := fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, recordID, zoneID)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		zoneLookupUrl,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set(headerContentType, mimeForm)

	res, err := c.client.Do(req)
	if err != nil || res == nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
	}

	return nil
}

func (c *tidyDNSClient) FindRecord(ctx context.Context, zoneID int, name string, rType RecordType) ([]*RecordInfo, error) {
	var records []recordList
	recordLookupUrl := fmt.Sprintf("%s/=/record?type=json&zone=%d&name=%s", c.baseURL, zoneID, name)
	err := c.getData(
		ctx,
		recordLookupUrl,
		&records,
	)
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
	recordMergeUrl := fmt.Sprintf("%s/=/record_merged?type=json&zone_id=%d&showall=1", c.baseURL, zoneID)
	err := c.getData(
		ctx,
		recordMergeUrl,
		&records,
	)
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
	recordLookupUrl := fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, zoneID, recordID)
	err := c.getData(
		ctx,
		recordLookupUrl,
		&record,
	)
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
	recordLookupUrl := fmt.Sprintf("%s/=/record/%d/%d", c.baseURL, recordID, zoneID)
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		recordLookupUrl,
		nil,
	)
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.username, c.password)

	res, err := c.client.Do(req)
	if err != nil || res == nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
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
	if err != nil || res == nil {
		return err
	}
	defer closeResponse(res)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(errorTidyDNS, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(value)
	if err != nil {
		return err
	}

	return nil
}

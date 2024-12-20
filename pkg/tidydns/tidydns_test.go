package tidydns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const subnetResponse = `[
  {
    "vlan_name": "netic-shared-k8s-utility01",
    "created_date": "2021-04-16 14:46:46",
    "modified_by": "mef",
    "vlan_id": 959,
    "loc_name": "internal",
    "subnet_nice_name": "10.68.0.128/26 - netic-shared-k8s-utility01-v4",
    "dhcp_failover": 1,
    "vlan": "534 netic-shared-k8s-utility01",
    "status": 0,
    "dhcp_active": 0,
    "location_id": 1,
    "description": null,
    "id": 1185,
    "modified_date": "2021-07-08 10:38:02",
    "name": "netic-shared-k8s-utility01-v4",
    "subnet": "10.68.0.128/26",
    "zone": "k8s.netic.dk",
    "customer_id": 0,
    "family": 4,
    "zone_id": 2861,
    "vmps_active": 0,
    "shared_network": null,
    "icons": "<i class=\"fa fa-circle-o fa-fw\" title=\"DHCP Deactive\"></i><i class=\"fa fa-circle-o fa-fw\" title=\"VMPS Deactive\"></i>",
    "vlan_no": 534
  }
]`

const freeIPResponse = `{
  "status": 0,
  "data": {
    "name_suggestion": "netic-shared-k8s-utility01-v4-134",
    "last_octet": "134",
    "ip_address": "10.68.0.134"
  }
}`

const createResponseV1 = `{
  "status": "0",
  "id": 30641,
  "subnet_id": 1185
}`

const createResponseV2 = `{
  "status": 0,
  "id": 30641,
  "subnet_id": 1185
}`

const readResponse = `{
  "brother_id": null,
  "id": 30641,
  "extra_ip": 0,
  "address_id": 63573,
  "subnet_id": 1185,
  "modified_date": "2021-07-08",
  "subnet": "10.68.0.128/26",
  "mac_addr": null,
  "destination": "10.68.0.134",
  "fqdn_namehelper": null,
  "vlan_id": 959,
  "icons": "<i class=\"fa fa-circle-o fa-fw\" title=\"DHCP Deactive\"></i><i class=\"fa fa-circle-o fa-fw\" title=\"VMPS Deactive\"></i><i class=\"fa fa-circle fa-fw\" title=\"DNS Active\"></i>",
  "description": "",
  "aliases": "",
  "type": 1,
  "dhcp_active": 0,
  "subnet_name": "netic-shared-k8s-utility01-v4",
  "ip_family": 4,
  "is_template": 0,
  "zone": "k8s.netic.dk",
  "name": "test-tal",
  "modified_by": "tal",
  "customer_id_inherited": 0,
  "customer_id": null,
  "vlan_name": "netic-shared-k8s-utility01",
  "star": "<i class=\"fa fa-star-o fa-lg\" title=\"Click to Set Interface as Template\"></i>",
  "vmps_active": 0,
  "vlan_no": 534,
  "record_id": 30641,
  "ip_address": "10.68.0.134",
  "status": 0,
  "fqdn_name": "test-tal.k8s.netic.dk",
  "seen_date": null,
  "fqdn": "test-tal.k8s.netic.dk",
  "location_id": 1,
  "dual_stack": "",
  "zone_id": 2861,
  "brother_destination": null
}`

const userCreateResponse = `{"data":{"id":144},"status":"0"}`
const userReadResponse = `{"modified_by":"jra-api-test","description":"Awesome test user","modified_date":"2024-12-03 14:17:22","username":"jra-test-user","auth_group":"User","name":"jra-test-user","epassword":"*****","passwd_changed_date":"2024-12-03 14:17:22","id":148,"groups":[{"groupname":"user","name":"User","notes":null,"id":2,"description":null}]}`
const userUpdateResponse = `{"status":"0","data":{"id":146}}`
const userDeleteResponse = `{"status":"0"}`

func TestGetSubnetIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "10.68.0.128/26", req.URL.Query().Get("subnet"))
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(subnetResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	ids, err := c.GetSubnetIDs(context.Background(), "10.68.0.128/26")
	assert.NoError(t, err)
	assert.Equal(t, 1185, ids.SubnetID)
	assert.Equal(t, 2861, ids.ZoneID)
	assert.Equal(t, 534, ids.VlanNo)
}

func TestGetFreeIP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "1185")
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(freeIPResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	ip, err := c.GetFreeIP(context.Background(), 1185)
	assert.NoError(t, err)
	assert.Equal(t, "10.68.0.134", ip)
}

func TestCreateDHCPInterfaceV1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		_, _ = rw.Write([]byte(createResponseV1))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	createInfo := CreateInfo{
		SubnetID:      1185,
		ZoneID:        2861,
		InterfaceIP:   "8.8.8.8",
		InterfaceName: "unittest",
	}
	id, err := c.CreateDHCPInterface(context.Background(), createInfo)
	assert.NoError(t, err)
	assert.Equal(t, 30641, id)
}

func TestCreateDHCPInterfaceV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		_, _ = rw.Write([]byte(createResponseV2))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	createInfo := CreateInfo{
		SubnetID:      1185,
		ZoneID:        2861,
		InterfaceIP:   "8.8.8.8",
		InterfaceName: "unittest",
	}
	id, err := c.CreateDHCPInterface(context.Background(), createInfo)
	assert.NoError(t, err)
	assert.Equal(t, 30641, id)
}

func TestReadDHCPInterface(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "30641", req.URL.Query().Get("id"))
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(readResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	info, err := c.ReadDHCPInterface(context.Background(), 30641)
	assert.NoError(t, err)
	assert.Equal(t, "10.68.0.134", info.InterfaceIP)
	assert.Equal(t, "test-tal", info.Interfacename)
}

func TestUpdateDHCPInterfaceName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "30641")
		assert.Equal(t, "POST", req.Method)
		_, _ = rw.Write([]byte(createResponseV1))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	id, err := c.UpdateDHCPInterfaceName(context.Background(), 30641, "test-tal-update")
	assert.NoError(t, err)
	assert.Equal(t, 30641, id)
}

func TestDeleteDHCPInterface(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "30641")
		assert.Equal(t, "DELETE", req.Method)
		_, _ = rw.Write([]byte(createResponseV1))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	err := c.DeleteDHCPInterface(context.Background(), 30641)
	assert.NoError(t, err)
}

func TestFindZoneID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Query().Get("name"), "hackerdays.trifork.dev")
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(zoneSearchResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	id, err := c.FindZoneID(context.Background(), "hackerdays.trifork.dev")
	assert.NoError(t, err)
	assert.Equal(t, 2926, id)
}

func TestCreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(readRecordListResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	createInfo := RecordInfo{
		Type:        RecordTypeA,
		Name:        "tal-test",
		Destination: "10.68.1.2",
		Location:    LocationID(0),
	}
	id, err := c.CreateRecord(context.Background(), 2861, createInfo)
	assert.NoError(t, err)
	assert.Equal(t, 64694, id)
}

func TestReadRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "2861")
		assert.Contains(t, req.URL.Path, "64694")
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(readRecordResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	info, err := c.ReadRecord(context.Background(), 2861, 64694)

	assert.NoError(t, err)

	assert.Equal(t, "10.68.1.2", info.Destination)
	assert.Equal(t, "tal-test", info.Name)
	assert.Equal(t, RecordTypeA, info.Type)
}

func TestDeleteRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Contains(t, req.URL.Path, "2861")
		assert.Contains(t, req.URL.Path, "64694")
		assert.Equal(t, "DELETE", req.Method)
		_, _ = rw.Write([]byte(createResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	err := c.DeleteRecord(context.Background(), 2861, 64694)
	assert.NoError(t, err)
}

func TestUpdateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		_, _ = rw.Write([]byte(readRecordListResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	createInfo := RecordInfo{
		Destination: "10.68.1.2",
		Location:    LocationID(0),
	}
	err := c.UpdateRecord(context.Background(), 2861, 64694, createInfo)
	assert.NoError(t, err)
}

func TestFindRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "prod1-api.trifork.shared", req.URL.Query().Get("name"))
		assert.Equal(t, "2861", req.URL.Query().Get("zone"))
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(findRecordResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	info, err := c.FindRecord(context.Background(), 2861, "prod1-api.trifork.shared", RecordTypeA)
	assert.NoError(t, err)
	assert.Equal(t, "prod1-api.trifork.shared", info[0].Name)
	assert.Equal(t, 65377, info[0].ID)
}

func TestListZones(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(listZonesResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	zones, err := c.ListZones(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, len(zones), 4)
}

func TestListRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "2861", req.URL.Query().Get("zone_id"))
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(listRecordsResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	records, err := c.ListRecords(context.Background(), 2861)
	assert.NoError(t, err)
	assert.Equal(t, len(records), 22)
}

func TestCreateInternalUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.NoError(t, req.ParseForm())
		assert.Equal(t, "/=/user/new", req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "test_user", req.PostForm.Get("username"))
		assert.Equal(t, "test_password", req.PostForm.Get("epassword"))
		assert.Equal(t, "test_password", req.PostForm.Get("epassword_verify"))
		assert.Equal(t, "0", req.PostForm.Get("change_password_on_first_login"))
		assert.Equal(t, "2", req.PostForm.Get("auth_group"))
		assert.Equal(t, "", req.PostForm.Get("user_allow"))
		_, _ = rw.Write([]byte(userCreateResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	id, err := c.CreateInternalUser(
		context.Background(),
		"test_user",
		"test_password",
		"description",
		false,
		AuthGroupUser,
		[]UserAllowID{},
	)
	assert.NoError(t, err)
	assert.Equal(t, id, UserID(144))
}

func TestGetInternalUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/=/user/144", req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		_, _ = rw.Write([]byte(userReadResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	userInfo, err := c.GetInternalUser(context.Background(), 144)
	assert.NoError(t, err)
	assert.Equal(t, userInfo.AuthGroup, AuthGroupUser)
	assert.Equal(t, userInfo.Description, "Awesome test user")
	assert.Equal(t, len(userInfo.Groups), 1)
	assert.Equal(t, userInfo.Groups[0].Id, 2)
	assert.Equal(t, userInfo.Groups[0].Name, "User")
	assert.Nil(t, userInfo.Groups[0].Notes)
	assert.Equal(t, userInfo.Groups[0].GroupName, "user")
	assert.Nil(t, userInfo.Groups[0].Description)
	assert.Equal(t, userInfo.Id, UserID(148))
	assert.Equal(t, userInfo.ModifiedBy, "jra-api-test")
	assert.Equal(t, userInfo.ModifiedDate, time.Date(2024, 12, 03, 14, 17, 22, 0, time.UTC))
	assert.Equal(t, userInfo.Name, "jra-test-user")
	assert.Equal(t, userInfo.PasswdChangedDate, time.Date(2024, 12, 03, 14, 17, 22, 0, time.UTC))
	assert.Equal(t, userInfo.Username, "jra-test-user")
}

func toPtr[T any](s T) *T {
	return &s
}

func TestUpdateInternalUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.NoError(t, req.ParseForm())
		assert.Equal(t, "/=/user/146", req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.False(t, req.PostForm.Has("username"))
		assert.Equal(t, "test_password", req.PostForm.Get("epassword"))
		assert.Equal(t, "test_password", req.PostForm.Get("epassword_verify"))
		assert.False(t, req.PostForm.Has("change_password_on_first_login"))
		assert.Equal(t, "2", req.PostForm.Get("auth_group"))
		assert.Equal(t, "", req.PostForm.Get("user_allow"))
		assert.Equal(t, "desc", req.PostForm.Get("description"))
		_, _ = rw.Write([]byte(userUpdateResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	err := c.UpdateInternalUser(
		context.Background(),
		UserID(146),
		toPtr("test_password"),
		toPtr("desc"),
		toPtr(AuthGroupUser),
		[]UserAllowID{},
	)
	assert.NoError(t, err)
}

func TestDeleteInternalUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/=/user/144", req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		_, _ = rw.Write([]byte(userDeleteResponse))
	}))
	defer server.Close()

	c := New(server.URL, "username", "password")
	err := c.DeleteInternalUser(context.Background(), UserID(144))
	assert.NoError(t, err)
}

const zoneSearchResponse = `[
  {
    "soa_record": null,
    "namehelper": null,
    "network": null,
    "customer_id": 0,
    "authoritative_log": "NXDOMAIN: ASKED(ns-cloud-e4.googledomains.com ns-cloud-e2.googledomains.com ns-cloud-e3.googledomains.com ns-cloud-e1.googledomains.com)",
    "class": null,
    "soa_slave_retry": null,
    "last_check_date": "2021-09-07 14:25:24",
    "soa_ttl": null,
    "autofill_template": null,
    "soa_slave_refresh": null,
    "serial": 8,
    "type_text": "regular",
    "dnssec_parent_state": -1,
    "dnssec_genkeys": 1,
    "update": 0,
    "server_state": 0,
    "masters": null,
    "provision_date": "2021-09-17 08:43:21",
    "soa_max_caching": null,
    "range_stop": null,
    "provision_log": null,
    "id": 2926,
    "created_date": "2021-09-07 14:25:22",
    "type_has_records": 1,
    "description": "Subdomain delegation for automated DNS @Hackerdays 2021",
    "allow_transfer": null,
    "modified_by": "tal",
    "type": 0,
    "alias_type": null,
    "dnssec_lastsign": null,
    "autofill_enable": null,
    "dnssec_enable": 0,
    "alias_id": null,
    "provision_state": 1,
    "authoritative_state": 3,
    "parent_id": null,
    "do_validate": 0,
    "forwarders": null,
    "dnssec_parent_log": null,
    "status": 0,
    "dnssec_monitor_enable": 0,
    "alias_name": null,
    "soa_contact": null,
    "soa_slave_expiration": null,
    "modified_date": "2021-09-17 08:43:19",
    "is_private": null,
    "inject_ns_enable": 1,
    "zone_name": "hackerdays.trifork.dev",
    "range_start": null,
    "name": "hackerdays.trifork.dev"
  }
]`

const createResponse = `{
  "status": "0",
  "id": 30641,
  "subnet_id": 1185
}`

const readRecordResponse = `{
  "zone_record": 0,
  "macro": 0,
  "nice_macro_name": null,
  "loc_name": "internal",
  "modified_date": "2021-08-18 12:53:50",
  "destination": "10.68.1.2",
  "ttl": 0,
  "status_text": "active",
  "zone_status": 0,
  "modified_by": "tal",
  "created_by": "tal",
  "status": 0,
  "type_name": "A",
  "namehelper": null,
  "macro_name": null,
  "tidy_record": 1,
  "zone_type": 0,
  "external_table": "tidy_record",
  "description": "Test A record creation",
  "location_id": 1,
  "port": null,
  "name": "tal-test",
  "created_date": "2021-08-18 12:53:50",
  "history": 0,
  "type": 0,
  "value": 0,
  "zone_id": 2861,
  "zone_name": "k8s.netic.dk",
  "customer_id": 0,
  "extra_ip": null,
  "active": 1,
  "coalesce_name": "tal-test",
  "weight": null,
  "id": 64694,
  "coalesce_macro_dest": "10.68.1.2",
  "record_fullname": "tal-test.k8s.netic.dk"
}`

const readRecordListResponse = `[  {
  "coalesce_name": "tal-test",
  "reverse_range_stop": null,
  "port": null,
  "weight": null,
  "description": null,
  "status_text": "active",
  "deleted": "",
  "reverse_range_start": null,
  "active": "1",
  "id": 64694,
  "reverse_network": null,
  "external_table": "tidy_record",
  "macro": 0,
  "zone_type": 0,
  "zone_name": "k8s.netic.dk",
  "zone_id": 2861,
  "namehelper": null,
  "name": "tal-test",
  "reverse_class": null,
  "coalesce_macro_dest": "10.68.1.2",
  "loc_name": "internal",
  "modified_by": "ash",
  "location_id": 1,
  "modified_date": "2021-04-19 07:14:04",
  "type": 0,
  "zone_record_grouping": "A",
  "destination": "10.68.1.2",
  "extra_ip": 0,
  "macro_name": null,
  "tidy_record": 0,
  "value": 0,
  "ttl": 0,
  "status": "0",
  "customer_id": 0,
  "zone_status": 0,
  "type_name": "A",
  "zone_record": 0,
  "history": 0,
  "nice_macro_name": null
},{
  "extra_ip": null,
  "active": 0,
  "loc_name": "all",
  "weight": null,
  "type_name": "NS",
  "reverse_class": null,
  "zone_status": "0",
  "destination": "a.ns.netic.dk.",
  "reverse_range_start": null,
  "customer_id": null,
  "macro": null,
  "reverse_range_stop": null,
  "name": ".",
  "zone_id": 2861,
  "tidy_record": 1,
  "ttl": 0,
  "modified_by": null,
  "status": -1,
  "macro_name": null,
  "nice_macro_name": null,
  "coalesce_name": ".",
  "reverse_network": null,
  "namehelper": null,
  "type": 4,
  "zone_record_grouping": 0,
  "deleted": 1,
  "zone_name": null,
  "modified_date": null,
  "zone_type": null,
  "location_id": null,
  "coalesce_macro_dest": null,
  "port": null,
  "status_text": null,
  "zone_record": 1,
  "value": null,
  "id": null,
  "history": null,
  "external_table": null,
  "description": "inherired from soa configuration"
}]`

const findRecordResponse = `[
  {
    "modified_date": "2021-09-13 10:06:07",
    "macro_name": null,
    "status": 0,
    "extra_ip": null,
    "name": "prod1-api.trifork.shared",
    "status_text": "active",
    "type_name": "A",
    "zone_name": "k8s.netic.dk",
    "active": 1,
    "zone_record": 0,
    "description": null,
    "type": 0,
    "modified_by": "api-terraform-shared-k8s",
    "destination": "77.243.49.187",
    "zone_status": 0,
    "created_by": "api-terraform-shared-k8s",
    "external_table": "tidy_record",
    "record_fullname": "prod1-api.trifork.shared.k8s.netic.dk",
    "id": 65377,
    "zone_type": 0,
    "coalesce_name": "prod1-api.trifork.shared",
    "created_date": "2021-09-13 10:06:07",
    "location_id": 0,
    "loc_name": "all",
    "port": null,
    "ttl": 3600,
    "tidy_record": 1,
    "value": 0,
    "namehelper": null,
    "zone_id": 2861,
    "customer_id": 0,
    "weight": null,
    "coalesce_macro_dest": "77.243.49.187",
    "macro": 0,
    "nice_macro_name": null,
    "history": 0
  }
]`

const listZonesResponse = `[
  {
    "id": 2926,
    "last_check_date": "2021-09-07 14:25:24",
    "masters": null,
    "alias_type": null,
    "server_state": 0,
    "authoritative_log": "NXDOMAIN: ASKED(ns-cloud-e4.googledomains.com ns-cloud-e2.googledomains.com ns-cloud-e3.googledomains.com ns-cloud-e1.googledomains.com)",
    "class": null,
    "forwarders": null,
    "soa_ttl": null,
    "is_private": null,
    "soa_slave_refresh": null,
    "inject_ns_enable": 1,
    "dnssec_lastsign": null,
    "zone_name": "hackerdays.trifork.dev",
    "soa_slave_expiration": null,
    "soa_max_caching": null,
    "dnssec_parent_state": -1,
    "provision_state": 1,
    "type": 0,
    "authoritative_state": 3,
    "network": null,
    "update": 0,
    "type_has_records": 1,
    "dnssec_monitor_enable": 0,
    "created_date": "2021-09-07 14:25:22",
    "dnssec_enable": 0,
    "range_stop": null,
    "dnssec_genkeys": 1,
    "type_text": "regular",
    "customer_id": 0,
    "autofill_enable": null,
    "name": "hackerdays.trifork.dev",
    "namehelper": null,
    "range_start": null,
    "soa_record": null,
    "autofill_template": null,
    "dnssec_parent_log": null,
    "status": 0,
    "description": "Subdomain delegation for automated DNS @Hackerdays 2021",
    "alias_id": null,
    "soa_slave_retry": null,
    "provision_log": null,
    "alias_name": null,
    "modified_date": "2021-10-01 15:59:44",
    "modified_by": "api-letsencrypt-shared-k8s",
    "allow_transfer": null,
    "provision_date": "2021-10-01 15:59:45",
    "soa_contact": null,
    "parent_id": null,
    "do_validate": 0,
    "serial": 502
  },
  {
    "provision_date": "2021-10-04 14:31:46",
    "allow_transfer": null,
    "serial": 624,
    "parent_id": 279,
    "do_validate": 0,
    "soa_contact": null,
    "provision_log": null,
    "soa_slave_retry": null,
    "alias_id": null,
    "modified_by": "api-terraform-shared-k8s",
    "modified_date": "2021-10-04 14:31:44",
    "alias_name": null,
    "dnssec_parent_log": null,
    "autofill_template": null,
    "soa_record": null,
    "range_start": null,
    "name": "k8s.netic.dk",
    "namehelper": null,
    "autofill_enable": null,
    "status": 0,
    "description": null,
    "created_date": "2021-04-16 14:43:04",
    "dnssec_monitor_enable": 0,
    "type_has_records": 1,
    "update": 0,
    "dnssec_genkeys": 1,
    "type_text": "regular",
    "customer_id": 0,
    "range_stop": null,
    "dnssec_enable": 0,
    "type": 0,
    "provision_state": 1,
    "dnssec_parent_state": -1,
    "soa_slave_expiration": null,
    "soa_max_caching": null,
    "authoritative_state": 1,
    "network": null,
    "inject_ns_enable": 1,
    "soa_slave_refresh": null,
    "is_private": null,
    "dnssec_lastsign": null,
    "zone_name": "k8s.netic.dk",
    "class": null,
    "server_state": 0,
    "authoritative_log": "NOERROR: ASKED(a.ns.netic.dk b.ns.netic.dk c.ns.netic.dk): NS(a.ns.netic.dk c.ns.netic.dk b.ns.netic.dk)",
    "alias_type": null,
    "masters": null,
    "soa_ttl": null,
    "forwarders": null,
    "id": 2861,
    "last_check_date": "2021-10-04 09:59:25"
  },
  {
    "alias_name": null,
    "modified_by": "api-letsencrypt-wsus",
    "modified_date": "2021-10-04 13:38:36",
    "soa_slave_retry": null,
    "alias_id": null,
    "provision_log": "",
    "do_validate": 0,
    "parent_id": null,
    "soa_contact": null,
    "serial": 17830,
    "provision_date": "2021-10-04 13:38:56",
    "allow_transfer": null,
    "dnssec_enable": 1,
    "type_text": "regular",
    "dnssec_genkeys": 0,
    "customer_id": 0,
    "range_stop": null,
    "update": 0,
    "created_date": "2004-09-20 15:24:56",
    "type_has_records": 1,
    "dnssec_monitor_enable": 1,
    "status": 0,
    "description": "Netic A/S, Aalborg, Denmark",
    "range_start": null,
    "namehelper": null,
    "name": "netic.dk",
    "autofill_enable": null,
    "dnssec_parent_log": "",
    "soa_record": null,
    "autofill_template": null,
    "zone_name": "netic.dk",
    "dnssec_lastsign": "2021-10-04 13:38:56",
    "is_private": null,
    "inject_ns_enable": 1,
    "soa_slave_refresh": null,
    "authoritative_state": 1,
    "network": null,
    "soa_slave_expiration": null,
    "soa_max_caching": null,
    "type": 0,
    "dnssec_parent_state": 1,
    "provision_state": 2,
    "last_check_date": "2021-10-04 07:58:18",
    "id": 279,
    "soa_ttl": null,
    "forwarders": null,
    "masters": null,
    "authoritative_log": "NOERROR: ASKED(a.ns.netic.dk b.ns.netic.dk c.ns.netic.dk): NS(a.ns.netic.dk b.ns.netic.dk c.ns.netic.dk)",
    "server_state": 0,
    "class": null,
    "alias_type": null
  },
  {
    "range_start": null,
    "name": "netic.eu",
    "autofill_enable": null,
    "namehelper": null,
    "dnssec_parent_log": null,
    "autofill_template": null,
    "soa_record": null,
    "description": "",
    "status": 0,
    "update": 0,
    "created_date": "2009-04-14 10:10:50",
    "type_has_records": 0,
    "dnssec_monitor_enable": 0,
    "dnssec_enable": 0,
    "dnssec_genkeys": 1,
    "type_text": "alias",
    "customer_id": 0,
    "range_stop": null,
    "provision_date": "2021-10-04 13:39:09",
    "allow_transfer": null,
    "parent_id": null,
    "do_validate": 0,
    "soa_contact": null,
    "serial": 12925,
    "soa_slave_retry": null,
    "alias_id": 279,
    "provision_log": "",
    "alias_name": "netic.dk",
    "modified_by": "mni",
    "modified_date": "2021-10-04 13:38:36",
    "masters": null,
    "class": null,
    "server_state": 0,
    "authoritative_log": "NOERROR: ASKED(b.ns.netic.dk a.ns.netic.dk c.ns.netic.dk): NS(b.ns.netic.dk c.ns.netic.dk a.ns.netic.dk)",
    "alias_type": 0,
    "soa_ttl": null,
    "forwarders": null,
    "id": 1180,
    "last_check_date": "2021-10-04 12:51:23",
    "soa_max_caching": null,
    "soa_slave_expiration": null,
    "type": 2,
    "provision_state": 2,
    "dnssec_parent_state": -1,
    "authoritative_state": 1,
    "network": null,
    "is_private": null,
    "inject_ns_enable": 1,
    "soa_slave_refresh": null,
    "dnssec_lastsign": null,
    "zone_name": "netic.eu"
  }
]`

const listRecordsResponse = `[
  {
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "macro_name": null,
    "zone_status": 0,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-hotrod-app/hotrod",
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-hotrod-app/hotrod",
    "external_table": "tidy_record",
    "zone_record_grouping": "TXT",
    "history": 0,
    "type": 5,
    "weight": null,
    "zone_type": 0,
    "location_id": 0,
    "customer_id": 0,
    "coalesce_name": "hotrod",
    "port": null,
    "tidy_record": 1,
    "name": "hotrod",
    "modified_date": "2021-09-23 10:54:03",
    "reverse_network": null,
    "value": 0,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "id": 65682,
    "reverse_range_start": null,
    "deleted": "",
    "type_name": "TXT",
    "status_text": "active",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "reverse_class": null,
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s"
  },
  {
    "coalesce_macro_dest": "77.243.49.244",
    "zone_status": 0,
    "macro_name": null,
    "zone_id": 2926,
    "zone_record": 1,
    "reverse_range_stop": null,
    "history": 1,
    "zone_record_grouping": 0,
    "external_table": "tidy_record",
    "destination": "77.243.49.244",
    "description": "Default ingress to prod2.trifork.dedicated.k8s.netic.dk",
    "location_id": 0,
    "zone_type": 0,
    "weight": null,
    "type": 0,
    "port": null,
    "tidy_record": 1,
    "customer_id": 0,
    "coalesce_name": ".",
    "value": 0,
    "reverse_network": null,
    "modified_date": "2021-09-17 08:43:10",
    "name": ".",
    "reverse_range_start": null,
    "id": 65293,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "deleted": "",
    "modified_by": "tal",
    "namehelper": null,
    "extra_ip": null,
    "nice_macro_name": null,
    "active": "1",
    "reverse_class": null,
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "status_text": "active",
    "type_name": "A"
  },
  {
    "reverse_range_start": null,
    "id": 65291,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "reverse_network": null,
    "value": 0,
    "modified_date": "2021-09-07 14:25:22",
    "name": ".",
    "extra_ip": null,
    "modified_by": "tal",
    "namehelper": null,
    "active": "1",
    "nice_macro_name": "<i class=\"fa fa-angle-double-right\"></i> NAMESERVER_C",
    "reverse_class": null,
    "status_text": "active",
    "type_name": "NS",
    "macro": 1,
    "zone_name": "hackerdays.trifork.dev",
    "deleted": "",
    "zone_record_grouping": 0,
    "history": 0,
    "external_table": "tidy_record",
    "description": "",
    "destination": "c.ns.netic.dk.",
    "macro_name": "$NAMESERVER_C$",
    "coalesce_macro_dest": "$NAMESERVER_C$",
    "zone_status": 0,
    "zone_id": 2926,
    "zone_record": 1,
    "reverse_range_stop": null,
    "port": null,
    "tidy_record": 1,
    "customer_id": 0,
    "coalesce_name": ".",
    "location_id": 0,
    "weight": null,
    "zone_type": 0,
    "type": 4
  },
  {
    "external_table": "tidy_record",
    "description": null,
    "destination": "b.ns.netic.dk.",
    "zone_record_grouping": 0,
    "history": 0,
    "zone_record": 1,
    "reverse_range_stop": null,
    "macro_name": "$NAMESERVER_B$",
    "zone_status": 0,
    "coalesce_macro_dest": "$NAMESERVER_B$",
    "zone_id": 2926,
    "port": null,
    "tidy_record": 1,
    "customer_id": 0,
    "coalesce_name": ".",
    "weight": null,
    "zone_type": 0,
    "type": 4,
    "location_id": 0,
    "id": 65294,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "reverse_range_start": null,
    "modified_date": "2021-09-07 14:25:22",
    "name": ".",
    "reverse_network": null,
    "value": 0,
    "reverse_class": null,
    "status_text": "active",
    "type_name": "NS",
    "macro": 1,
    "zone_name": "hackerdays.trifork.dev",
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "tal",
    "active": "1",
    "nice_macro_name": "<i class=\"fa fa-angle-double-right\"></i> NAMESERVER_B",
    "deleted": ""
  },
  {
    "name": "angular",
    "modified_date": "2021-09-24 13:21:02",
    "reverse_network": null,
    "value": 0,
    "loc_name": "all",
    "status": "0",
    "ttl": 0,
    "id": 66727,
    "reverse_range_start": null,
    "deleted": "",
    "type_name": "TXT",
    "status_text": "active",
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "reverse_class": null,
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "macro_name": null,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/angular-example-app/angular-example",
    "zone_status": 0,
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/angular-example-app/angular-example",
    "external_table": "tidy_record",
    "zone_record_grouping": "TXT",
    "history": 0,
    "type": 5,
    "weight": null,
    "zone_type": 0,
    "location_id": 0,
    "coalesce_name": "angular",
    "customer_id": 0,
    "port": null,
    "tidy_record": 1
  },
  {
    "deleted": "",
    "nice_macro_name": null,
    "active": "1",
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "extra_ip": null,
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "status_text": "active",
    "type_name": "A",
    "reverse_class": null,
    "value": 0,
    "reverse_network": null,
    "modified_date": "2021-10-01 11:29:56",
    "name": "dev.demo",
    "reverse_range_start": null,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "id": 66875,
    "location_id": 0,
    "type": 0,
    "zone_type": 0,
    "weight": null,
    "customer_id": 0,
    "coalesce_name": "dev.demo",
    "port": null,
    "tidy_record": 1,
    "zone_id": 2926,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "macro_name": null,
    "reverse_range_stop": null,
    "zone_record": 0,
    "history": 0,
    "zone_record_grouping": "A",
    "destination": "77.243.49.244",
    "description": null,
    "external_table": "tidy_record"
  },
  {
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "id": 66733,
    "reverse_range_start": null,
    "name": "grafana",
    "modified_date": "2021-09-27 12:52:24",
    "reverse_network": null,
    "value": 0,
    "status_text": "active",
    "type_name": "TXT",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "reverse_class": null,
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "deleted": "",
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/trifork-monitoring-system/grafana",
    "external_table": "tidy_record",
    "zone_record_grouping": "TXT",
    "history": 0,
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "macro_name": null,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/trifork-monitoring-system/grafana",
    "zone_status": 0,
    "customer_id": 0,
    "coalesce_name": "grafana",
    "port": null,
    "tidy_record": 1,
    "type": 5,
    "weight": null,
    "zone_type": 0,
    "location_id": 0
  },
  {
    "zone_record": 0,
    "reverse_range_stop": null,
    "macro_name": null,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "zone_id": 2926,
    "external_table": "tidy_record",
    "description": null,
    "destination": "77.243.49.244",
    "zone_record_grouping": "A",
    "history": 0,
    "weight": null,
    "zone_type": 0,
    "type": 0,
    "location_id": 0,
    "tidy_record": 1,
    "port": null,
    "customer_id": 0,
    "coalesce_name": "hotrod",
    "modified_date": "2021-09-23 10:54:01",
    "name": "hotrod",
    "reverse_network": null,
    "value": 0,
    "id": 65681,
    "loc_name": "all",
    "status": "0",
    "ttl": 0,
    "reverse_range_start": null,
    "deleted": "",
    "reverse_class": null,
    "type_name": "A",
    "status_text": "active",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "extra_ip": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "active": "1",
    "nice_macro_name": null
  },
  {
    "tidy_record": 1,
    "port": null,
    "coalesce_name": "sonarqube",
    "customer_id": 0,
    "zone_type": 0,
    "weight": null,
    "type": 0,
    "location_id": 0,
    "external_table": "tidy_record",
    "destination": "77.243.49.244",
    "description": null,
    "history": 0,
    "zone_record_grouping": "A",
    "zone_record": 0,
    "reverse_range_stop": null,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "macro_name": null,
    "zone_id": 2926,
    "reverse_class": null,
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "type_name": "A",
    "status_text": "active",
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "extra_ip": null,
    "nice_macro_name": null,
    "active": "1",
    "deleted": "",
    "id": 65701,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "reverse_range_start": null,
    "modified_date": "2021-09-23 13:17:46",
    "name": "sonarqube",
    "value": 0,
    "reverse_network": null
  },
  {
    "weight": null,
    "zone_type": 0,
    "type": 5,
    "location_id": 0,
    "port": null,
    "tidy_record": 1,
    "customer_id": 0,
    "coalesce_name": "dev.demo",
    "zone_record": 0,
    "reverse_range_stop": null,
    "macro_name": null,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-demo-dev/hotrod",
    "zone_status": 0,
    "zone_id": 2926,
    "external_table": "tidy_record",
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-demo-dev/hotrod",
    "zone_record_grouping": "TXT",
    "history": 0,
    "deleted": "",
    "reverse_class": null,
    "type_name": "TXT",
    "status_text": "active",
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "active": "1",
    "nice_macro_name": null,
    "modified_date": "2021-10-01 11:29:56",
    "name": "dev.demo",
    "reverse_network": null,
    "value": 0,
    "id": 66876,
    "loc_name": "all",
    "status": "0",
    "ttl": 0,
    "reverse_range_start": null
  },
  {
    "type": 0,
    "weight": null,
    "zone_type": 0,
    "location_id": 0,
    "coalesce_name": "angular",
    "customer_id": 0,
    "port": null,
    "tidy_record": 1,
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "macro_name": null,
    "coalesce_macro_dest": "77.243.49.244",
    "zone_status": 0,
    "description": null,
    "destination": "77.243.49.244",
    "external_table": "tidy_record",
    "zone_record_grouping": "A",
    "history": 0,
    "deleted": "",
    "type_name": "A",
    "status_text": "active",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "reverse_class": null,
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "modified_date": "2021-09-24 13:21:02",
    "name": "angular",
    "reverse_network": null,
    "value": 0,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "id": 66726,
    "reverse_range_start": null
  },
  {
    "value": 0,
    "reverse_network": null,
    "modified_date": "2021-09-30 11:32:57",
    "name": "springboot",
    "reverse_range_start": null,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "id": 66865,
    "deleted": "",
    "nice_macro_name": null,
    "active": "1",
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "extra_ip": null,
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "type_name": "TXT",
    "status_text": "active",
    "reverse_class": null,
    "zone_id": 2926,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/spring-boot-example-app/spring-boot-example",
    "zone_status": 0,
    "macro_name": null,
    "reverse_range_stop": null,
    "zone_record": 0,
    "history": 0,
    "zone_record_grouping": "TXT",
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/spring-boot-example-app/spring-boot-example",
    "description": null,
    "external_table": "tidy_record",
    "location_id": 0,
    "type": 5,
    "zone_type": 0,
    "weight": null,
    "coalesce_name": "springboot",
    "customer_id": 0,
    "port": null,
    "tidy_record": 1
  },
  {
    "reverse_network": null,
    "value": 0,
    "modified_date": "2021-09-23 13:17:46",
    "name": "sonarqube",
    "reverse_range_start": null,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "id": 65702,
    "deleted": "",
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "status_text": "active",
    "type_name": "TXT",
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "reverse_class": null,
    "zone_id": 2926,
    "macro_name": null,
    "zone_status": 0,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/sonarqube/sonarqube",
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_record_grouping": "TXT",
    "history": 0,
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/sonarqube/sonarqube",
    "external_table": "tidy_record",
    "location_id": 0,
    "type": 5,
    "weight": null,
    "zone_type": 0,
    "customer_id": 0,
    "coalesce_name": "sonarqube",
    "port": null,
    "tidy_record": 1
  },
  {
    "modified_date": "2021-09-30 10:56:37",
    "name": "podinfo",
    "reverse_network": null,
    "value": 0,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "id": 66862,
    "reverse_range_start": null,
    "deleted": "",
    "status_text": "active",
    "type_name": "TXT",
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "reverse_class": null,
    "active": "1",
    "nice_macro_name": null,
    "extra_ip": null,
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "macro_name": null,
    "zone_status": 0,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/podinfo-example-app/podinfo",
    "description": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/podinfo-example-app/podinfo",
    "external_table": "tidy_record",
    "zone_record_grouping": "TXT",
    "history": 0,
    "type": 5,
    "weight": null,
    "zone_type": 0,
    "location_id": 0,
    "coalesce_name": "podinfo",
    "customer_id": 0,
    "port": null,
    "tidy_record": 1
  },
  {
    "port": null,
    "tidy_record": 1,
    "coalesce_name": "dev-demo",
    "customer_id": 0,
    "zone_type": 0,
    "weight": null,
    "type": 5,
    "location_id": 0,
    "external_table": "tidy_record",
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-demo-dev/hotrod",
    "description": null,
    "history": 0,
    "zone_record_grouping": "TXT",
    "zone_record": 0,
    "reverse_range_stop": null,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/hackerdays-demo-dev/hotrod",
    "zone_status": 0,
    "macro_name": null,
    "zone_id": 2926,
    "reverse_class": null,
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "status_text": "active",
    "type_name": "TXT",
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "extra_ip": null,
    "nice_macro_name": null,
    "active": "1",
    "deleted": "",
    "id": 66879,
    "status": "0",
    "ttl": 0,
    "loc_name": "all",
    "reverse_range_start": null,
    "name": "dev-demo",
    "modified_date": "2021-10-01 11:30:56",
    "value": 0,
    "reverse_network": null
  },
  {
    "reverse_range_stop": null,
    "zone_record": 0,
    "zone_id": 2926,
    "coalesce_macro_dest": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/pvinddata/pvs-app-deny-actuator",
    "zone_status": 0,
    "macro_name": null,
    "destination": "heritage=external-dns,external-dns/owner=prod2.trifork,external-dns/resource=ingress/pvinddata/pvs-app-deny-actuator",
    "description": null,
    "external_table": "tidy_record",
    "history": 0,
    "zone_record_grouping": "TXT",
    "type": 5,
    "zone_type": 0,
    "weight": null,
    "location_id": 0,
    "customer_id": 0,
    "coalesce_name": "pvi",
    "tidy_record": 1,
    "port": null,
    "modified_date": "2021-10-01 15:59:29",
    "name": "pvi",
    "value": 0,
    "reverse_network": null,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "id": 66884,
    "reverse_range_start": null,
    "deleted": "",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "status_text": "active",
    "type_name": "TXT",
    "reverse_class": null,
    "nice_macro_name": null,
    "active": "1",
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "extra_ip": null
  },
  {
    "external_table": "tidy_record",
    "destination": "77.243.49.244",
    "description": null,
    "history": 0,
    "zone_record_grouping": "A",
    "zone_record": 0,
    "reverse_range_stop": null,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "macro_name": null,
    "zone_id": 2926,
    "port": null,
    "tidy_record": 1,
    "customer_id": 0,
    "coalesce_name": "pvi",
    "zone_type": 0,
    "weight": null,
    "type": 0,
    "location_id": 0,
    "id": 66883,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "reverse_range_start": null,
    "modified_date": "2021-10-01 15:59:29",
    "name": "pvi",
    "value": 0,
    "reverse_network": null,
    "reverse_class": null,
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "status_text": "active",
    "type_name": "A",
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "extra_ip": null,
    "nice_macro_name": null,
    "active": "1",
    "deleted": ""
  },
  {
    "port": null,
    "tidy_record": 1,
    "coalesce_name": "podinfo",
    "customer_id": 0,
    "weight": null,
    "zone_type": 0,
    "type": 0,
    "location_id": 0,
    "external_table": "tidy_record",
    "description": null,
    "destination": "77.243.49.244",
    "zone_record_grouping": "A",
    "history": 0,
    "zone_record": 0,
    "reverse_range_stop": null,
    "macro_name": null,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "zone_id": 2926,
    "reverse_class": null,
    "status_text": "active",
    "type_name": "A",
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "extra_ip": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "active": "1",
    "nice_macro_name": null,
    "deleted": "",
    "id": 66861,
    "loc_name": "all",
    "ttl": 0,
    "status": "0",
    "reverse_range_start": null,
    "name": "podinfo",
    "modified_date": "2021-09-30 10:56:37",
    "reverse_network": null,
    "value": 0
  },
  {
    "history": 0,
    "zone_record_grouping": "A",
    "destination": "77.243.49.244",
    "description": null,
    "external_table": "tidy_record",
    "zone_id": 2926,
    "coalesce_macro_dest": "77.243.49.244",
    "zone_status": 0,
    "macro_name": null,
    "reverse_range_stop": null,
    "zone_record": 0,
    "coalesce_name": "dev-demo",
    "customer_id": 0,
    "port": null,
    "tidy_record": 1,
    "location_id": 0,
    "type": 0,
    "zone_type": 0,
    "weight": null,
    "reverse_range_start": null,
    "status": "0",
    "ttl": 0,
    "loc_name": "all",
    "id": 66878,
    "value": 0,
    "reverse_network": null,
    "name": "dev-demo",
    "modified_date": "2021-10-01 11:30:56",
    "nice_macro_name": null,
    "active": "1",
    "namehelper": null,
    "modified_by": "api-letsencrypt-shared-k8s",
    "extra_ip": null,
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "type_name": "A",
    "status_text": "active",
    "reverse_class": null,
    "deleted": ""
  },
  {
    "reverse_range_start": null,
    "id": 66864,
    "ttl": 0,
    "status": "0",
    "loc_name": "all",
    "value": 0,
    "reverse_network": null,
    "modified_date": "2021-09-30 11:32:57",
    "name": "springboot",
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "extra_ip": null,
    "nice_macro_name": null,
    "active": "1",
    "reverse_class": null,
    "zone_name": "hackerdays.trifork.dev",
    "macro": 0,
    "type_name": "A",
    "status_text": "active",
    "deleted": "",
    "history": 0,
    "zone_record_grouping": "A",
    "external_table": "tidy_record",
    "destination": "77.243.49.244",
    "description": null,
    "coalesce_macro_dest": "77.243.49.244",
    "zone_status": 0,
    "macro_name": null,
    "zone_id": 2926,
    "zone_record": 0,
    "reverse_range_stop": null,
    "tidy_record": 1,
    "port": null,
    "customer_id": 0,
    "coalesce_name": "springboot",
    "location_id": 0,
    "zone_type": 0,
    "weight": null,
    "type": 0
  },
  {
    "location_id": 0,
    "type": 0,
    "zone_type": 0,
    "weight": null,
    "customer_id": 0,
    "coalesce_name": "grafana",
    "tidy_record": 1,
    "port": null,
    "zone_id": 2926,
    "zone_status": 0,
    "coalesce_macro_dest": "77.243.49.244",
    "macro_name": null,
    "reverse_range_stop": null,
    "zone_record": 0,
    "history": 0,
    "zone_record_grouping": "A",
    "destination": "77.243.49.244",
    "description": null,
    "external_table": "tidy_record",
    "deleted": "",
    "nice_macro_name": null,
    "active": "1",
    "modified_by": "api-letsencrypt-shared-k8s",
    "namehelper": null,
    "extra_ip": null,
    "macro": 0,
    "zone_name": "hackerdays.trifork.dev",
    "status_text": "active",
    "type_name": "A",
    "reverse_class": null,
    "value": 0,
    "reverse_network": null,
    "name": "grafana",
    "modified_date": "2021-09-27 12:52:24",
    "reverse_range_start": null,
    "status": "0",
    "ttl": 0,
    "loc_name": "all",
    "id": 66732
  },
  {
    "coalesce_name": ".",
    "customer_id": null,
    "port": null,
    "tidy_record": 1,
    "type": 4,
    "zone_type": null,
    "weight": null,
    "location_id": null,
    "destination": "a.ns.netic.dk.",
    "description": "inherired from soa configuration",
    "external_table": null,
    "history": null,
    "zone_record_grouping": 0,
    "reverse_range_stop": null,
    "zone_record": 1,
    "zone_id": 2926,
    "coalesce_macro_dest": null,
    "zone_status": "0",
    "macro_name": null,
    "zone_name": null,
    "macro": null,
    "type_name": "NS",
    "status_text": null,
    "reverse_class": null,
    "nice_macro_name": null,
    "active": 0,
    "modified_by": null,
    "namehelper": null,
    "extra_ip": null,
    "deleted": 1,
    "status": -1,
    "ttl": 0,
    "loc_name": "all",
    "id": null,
    "reverse_range_start": null,
    "name": ".",
    "modified_date": null,
    "value": null,
    "reverse_network": null
  }
]`

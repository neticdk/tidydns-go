package tidydns

type dhcpSubnet struct {
	ID         int `json:"id"`
	VlanId     int `json:"vlan_id"`
	VlanNo     int `json:"vlan_no"`
	ZoneID     int `json:"zone_id"`
	LocationID int `json:"location_id"`
}

type dhcpFreeIP struct {
	Status int            `json:"status"`
	Data   dhcpFreeIPData `json:"data"`
}

type dhcpFreeIPData struct {
	IPAddress string `json:"ip_address"`
}

type interfaceCreate struct {
	Status   interface{} `json:"status"`
	ID       int         `json:"id"`
	SubnetID int         `json:"subnet_id"`
}

type interfaceRead struct {
	Name        string `json:"name"`
	Destination string `json:"destination"`
}

type zoneInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type recordRead struct {
	ID          int          `json:"id"`
	Type        RecordType   `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Destination string       `json:"destination"`
	TTL         int          `json:"ttl"`
	Status      RecordStatus `json:"status"`
	Location    LocationID   `json:"location_id"`
}

type recordList struct {
	ID          int         `json:"id"`
	Type        RecordType  `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Destination string      `json:"destination"`
	TTL         int         `json:"ttl"`
	Status      interface{} `json:"status"`
	Location    LocationID  `json:"location_id"`
}

type userCreate struct {
	Data struct {
		Id int `json:"id"`
	} `json:"data"`
	Status string `json:"status"`
}

type userRead struct {
	ModifiedBy        string      `json:"modified_by"`
	Description       string      `json:"description"`
	ModifiedDate      string      `json:"modified_date"`
	Username          string      `json:"username"`
	AuthGroup         string      `json:"auth_group"`
	Name              string      `json:"name"`
	Epassword         string      `json:"epassword"`
	PasswdChangedDate string      `json:"passwd_changed_date"`
	Id                int         `json:"id"`
	Groups            []userGroup `json:"groups"`
}

type userGroup struct {
	GroupName   string  `json:"groupname"`
	Name        string  `json:"name"`
	Notes       *string `json:"notes,omitempty"`
	Id          int     `json:"id"`
	Description *string `json:"description,omitempty"`
}

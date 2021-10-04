package tidydns

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

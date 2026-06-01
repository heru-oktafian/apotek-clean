package models

// Response merepresentasikan format / struktur respons JSON standar
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ResponseGetAll merepresentasikan format / struktur respons JSON standar untuk pengambilan semua data
type ResponseGetAll struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	Search      string      `json:"search"`
	TotalItems  int         `json:"total_items"`
	CurrentPage int         `json:"current_page"`
	TotalPages  int         `json:"total_pages"`
	PerPage     int         `json:"per_page"`
	Data        interface{} `json:"data"`
}

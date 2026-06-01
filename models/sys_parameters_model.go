package models

// Struktur untuk menerima input dari body request
type RequestBody struct {
	Page   int    `json:"page" query:"page" form:"page"`
	Search string `json:"search" query:"search" form:"search"`
	Month  string `json:"month" query:"month" form:"month"`
}

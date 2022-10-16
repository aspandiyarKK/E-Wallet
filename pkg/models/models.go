package models

type TransactionQueryParams struct {
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
	Sort   string `json:"sort"`
	Desc   bool   `json:"desc"`
}

package models

type Statistics struct {
	Likes     int `json:"likes"`
	ViewCount int `json:"viewCount"`
	Contacts  int `json:"contacts"`
}

type CreateAdvertisementRequest struct {
	SellerID   int        `json:"sellerID"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
}

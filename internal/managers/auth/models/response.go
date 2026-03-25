package models

// Ответ от сервера, который описан в документации - первый позитивный тест проверяет, что возвращается именно он
type CreateAdvertisementResponseDocumentation struct {
	ID         string     `json:"id"`
	SellerID   int        `json:"sellerId"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
	CreatedAt  string     `json:"createdAt"`
}

// Так как в ходе первого позитивного теста было выявлено, что отет от сервера отличается от того, что описан в документации
// поэтому создаем структуру реального ответа от сервера
type CreateAdvertisementResponseServer struct {
	Status string `json:"status"`
}

type GetAdvertisementResponse struct {
	ID         string     `json:"id"`
	SellerID   int        `json:"sellerId"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
	CreatedAt  string     `json:"createdAt"`
}

type GetAdvertisementResponseList []GetAdvertisementResponse

type BadResponse struct {
	Result struct {
		Message  string                 `json:"message"`
		Messages map[string]interface{} `json:"messages"`
	} `json:"result"`
	Status string `json:"status"`
}

type NotFoundResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

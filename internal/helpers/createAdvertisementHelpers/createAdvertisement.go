package createAdvertisementHelpers

import (
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/utils"
)

func CreateTestAdvertisementHelper() *models.CreateAdvertisementRequest {
	return &models.CreateAdvertisementRequest{
		SellerID: utils.RandomInt(111111, 999999),
		Name:     utils.RandomString(10),
		Price:    utils.RandomInt(1, 100000),
		Statistics: models.Statistics{
			Likes:     utils.RandomInt(1, 100),
			ViewCount: utils.RandomInt(1, 100),
			Contacts:  utils.RandomInt(1, 100),
		},
	}
}

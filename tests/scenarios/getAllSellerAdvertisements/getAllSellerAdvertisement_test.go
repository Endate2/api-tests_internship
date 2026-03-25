package getAllSellerAdvertisements

import (
	"api-tests-template/internal/helpers/checkDeletedAdvertisementHelper"
	"api-tests-template/internal/helpers/createAdvertisementHelpers"
	"api-tests-template/internal/helpers/parseHelpers"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/createAdvertisement"
	"api-tests-template/internal/managers/deleteAdvertisement"
	"api-tests-template/internal/managers/getAdvertisementsBySellerId"
	"api-tests-template/internal/utils"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	base "api-tests-template/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	createdAds []string
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
}

func (s *TestSuite) TearDownTest() {
	t := s.T()
	t.Logf("TearDownTest: пытаемся удалить объявления: %v", s.createdAds)

	for _, id := range s.createdAds {
		deleteResponse := deleteAdvertisement.DeleteAdvertisement(t, http.StatusOK, id)
		t.Logf("Delete response for %s: %s", id, deleteResponse)

		isDeleted := checkDeletedAdvertisementHelper.VerifyAdvertisementDeleted(t, id)
		if isDeleted {
			t.Logf("Объявление %s успешно удалено", id)
		} else {
			t.Logf("Объявление %s не было удалено", id)
		}
	}

	s.createdAds = nil
	base.TearDownSuite()
}

// createAdsForSeller создаёт count объявлений с заданным sellerID.
// Возвращает срез запросов и срез созданных ID в том же порядке.
// Все созданные ID регистрируются на удаление после теста.
func (s *TestSuite) createAdsForSeller(sellerID int, count int) ([]*models.CreateAdvertisementRequest, []string) {
	advertisements := make([]*models.CreateAdvertisementRequest, count)
	createdIDs := make([]string, count)

	for i := 0; i < count; i++ {
		ad := createAdvertisementHelpers.CreateTestAdvertisementHelper()
		ad.SellerID = sellerID
		advertisements[i] = ad

		_, responseBody := createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *ad)

		var response models.CreateAdvertisementResponseServer
		err := json.Unmarshal([]byte(responseBody), &response)
		require.NoError(s.T(), err, "Ошибка парсинга ответа при создании объявления %d", i+1)

		id := parseHelpers.ParseIDFromStatus(response.Status)
		require.NotEmpty(s.T(), id, "ID объявления %d не должен быть пустым", i+1)

		createdIDs[i] = id
		s.createdAds = append(s.createdAds, id)

		s.T().Logf("Создано объявление %d: ID=%s, Name=%s, Price=%d, SellerID=%d",
			i+1, id, ad.Name, ad.Price, ad.SellerID)
	}

	return advertisements, createdIDs
}

// assertBadResponse парсит тело ответа как BadResponse и проверяет код ошибки и сообщение.
func (s *TestSuite) assertBadResponse(responseBody string, expectedStatus string, expectedMessage string) {
	var receivedAnswer models.BadResponse
	err := json.Unmarshal([]byte(responseBody), &receivedAnswer)
	require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

	assert.Equal(s.T(), expectedStatus, receivedAnswer.Status,
		"Статус должен быть '%s', получено: %s", expectedStatus, receivedAnswer.Status)

	assert.Equal(s.T(), expectedMessage, receivedAnswer.Result.Message,
		"Некорректное сообщение об ошибке")
}

// Позитивный: три созданных объявления с одним sellerID должны присутствовать в ответе с корректными данными
func (s *TestSuite) TestGetAdvertisementsBySellerId_Success() {
	sellerID := utils.RandomInt(111111, 999999)

	var advertisements []*models.CreateAdvertisementRequest
	var createdIDs []string

	s.Run("Создание трёх объявлений с одинаковым sellerID", func() {
		advertisements, createdIDs = s.createAdsForSeller(sellerID, 3)
	})

	var receivedAdvertisements models.GetAdvertisementResponseList

	s.Run("Получение объявлений по sellerID", func() {
		_, getResponseBody := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)

		err := json.Unmarshal([]byte(getResponseBody), &receivedAdvertisements)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")
	})

	s.Run("Проверка что получено ровно 3 объявления", func() {
		assert.Equal(s.T(), 3, len(receivedAdvertisements),
			"Должно быть получено 3 объявления, получено: %d", len(receivedAdvertisements))
	})

	s.Run("Проверка данных каждого из трёх объявлений", func() {
		receivedMap := make(map[string]models.GetAdvertisementResponse)
		for _, ad := range receivedAdvertisements {
			receivedMap[ad.ID] = ad
		}

		for i, createdID := range createdIDs {
			expectedAd := advertisements[i]

			receivedAd, exists := receivedMap[createdID]
			require.True(s.T(), exists, "Объявление с ID %s не найдено в ответе", createdID)

			assert.Equal(s.T(), createdID, receivedAd.ID,
				"ID объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.SellerID, receivedAd.SellerID,
				"SellerID объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.Name, receivedAd.Name,
				"Name объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.Price, receivedAd.Price,
				"Price объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.Statistics.Likes, receivedAd.Statistics.Likes,
				"Statistics.Likes объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.Statistics.ViewCount, receivedAd.Statistics.ViewCount,
				"Statistics.ViewCount объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), expectedAd.Statistics.Contacts, receivedAd.Statistics.Contacts,
				"Statistics.Contacts объявления %d должен совпадать", i+1)
			assert.NotEmpty(s.T(), receivedAd.CreatedAt,
				"CreatedAt объявления %d не должен быть пустым", i+1)

			s.T().Logf("Проверено объявление %d: ID=%s, Name=%s, Price=%d",
				i+1, receivedAd.ID, receivedAd.Name, receivedAd.Price)
		}
	})
}

// Позитивный: все объявления в ответе должны иметь sellerID, совпадающий с запрошенным
func (s *TestSuite) TestGetAdvertisementsBySellerId_AllItemsHaveCorrectSellerID() {
	sellerID := utils.RandomInt(111111, 999999)

	s.Run("Создание трёх объявлений с одинаковым sellerID", func() {
		s.createAdsForSeller(sellerID, 3)
	})

	var receivedAdvertisements models.GetAdvertisementResponseList

	s.Run("Получение объявлений по sellerID", func() {
		_, getResponseBody := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)

		s.T().Logf("Response body: %s", getResponseBody)

		err := json.Unmarshal([]byte(getResponseBody), &receivedAdvertisements)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")

		s.T().Logf("Количество полученных объявлений: %d", len(receivedAdvertisements))
	})

	s.Run("Проверка что каждое объявление в ответе содержит корректный sellerID", func() {
		require.NotEmpty(s.T(), receivedAdvertisements, "Ответ не должен быть пустым")

		for i, ad := range receivedAdvertisements {
			assert.Equal(s.T(), sellerID, ad.SellerID,
				"Объявление %d (ID: %s) должно иметь sellerID = %d, получено: %d",
				i+1, ad.ID, sellerID, ad.SellerID)
		}
	})
}

// Негативный: запрос с sellerID, для которого нет объявлений, должен возвращать пустой массив
func (s *TestSuite) TestGetAdvertisementsBySellerId_WithNotRealSellerId() {
	sellerID := utils.RandomInt(111111, 999999)

	var receivedAdvertisements models.GetAdvertisementResponseList

	s.Run("Получение объявлений по несуществующему sellerID", func() {
		_, getResponseBody := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)

		s.T().Logf("Response body: %s", getResponseBody)

		err := json.Unmarshal([]byte(getResponseBody), &receivedAdvertisements)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")

		s.T().Logf("Количество полученных объявлений: %d", len(receivedAdvertisements))
	})

	s.Run("Проверка что ответ содержит пустой массив", func() {
		assert.Empty(s.T(), receivedAdvertisements,
			"Для несуществующего sellerID должен возвращаться пустой массив")
		assert.Equal(s.T(), 0, len(receivedAdvertisements),
			"Количество объявлений должно быть 0")
	})
}

// Негативный: sellerID с отрицательным значением должен возвращать 400
func (s *TestSuite) TestGetAdvertisementsBySellerId_WithNegativeSellerId() {
	sellerID := utils.RandomInt(-100, -50)

	var getResponseBody string

	s.Run("Получение объявлений с отрицательным sellerID", func() {
		_, getResponseBody = getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusBadRequest,
			sellerID,
		)

		s.T().Logf("Response body: %s", getResponseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(getResponseBody, "400", "передан некорректный идентификатор продавца")
	})
}

// Негативный: дробное число в качестве sellerID должно возвращать 400
func (s *TestSuite) TestGetAdvertisementsBySellerId_WithFractionalSellerId() {
	sellerID := float64(utils.RandomInt(1, 98)) / 99

	var getResponseBody string

	s.Run("Получение объявлений с дробным sellerID", func() {
		_, getResponseBody = getAdvertisementsBySellerId.GetAdvertisementsBySellerIdFloat(
			s.T(),
			http.StatusBadRequest,
			sellerID,
		)

		s.T().Logf("Response body: %s", getResponseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(getResponseBody, "400", "передан некорректный идентификатор продавца")
	})
}

// Корнер-кейс: повторные запросы с одним и тем же sellerID должны возвращать идентичные данные (идемпотентность)
func (s *TestSuite) TestGetAdvertisementsBySellerId_Idempotency() {
	sellerID := utils.RandomInt(111111, 999999)

	s.Run("Создание трёх объявлений с одинаковым sellerID", func() {
		s.createAdsForSeller(sellerID, 3)
	})

	var firstResponse, secondResponse models.GetAdvertisementResponseList

	s.Run("Два последовательных GET-запроса с одним и тем же sellerID", func() {
		_, firstBody := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)
		_, secondBody := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)

		errFirst := json.Unmarshal([]byte(firstBody), &firstResponse)
		errSecond := json.Unmarshal([]byte(secondBody), &secondResponse)
		require.NoError(s.T(), errFirst, "Ошибка парсинга первого ответа")
		require.NoError(s.T(), errSecond, "Ошибка парсинга второго ответа")

		require.NotEmpty(s.T(), firstResponse, "Первый ответ не должен быть пустым")
		require.NotEmpty(s.T(), secondResponse, "Второй ответ не должен быть пустым")
	})

	s.Run("Проверка что оба ответа содержат одинаковое количество объявлений", func() {
		assert.Equal(s.T(), len(firstResponse), len(secondResponse),
			"Количество объявлений в обоих ответах должно совпадать: первый=%d, второй=%d",
			len(firstResponse), len(secondResponse))
	})

	s.Run("Проверка что данные каждого объявления совпадают в обоих ответах", func() {
		secondMap := make(map[string]models.GetAdvertisementResponse)
		for _, ad := range secondResponse {
			secondMap[ad.ID] = ad
		}

		for i, ad := range firstResponse {
			corresponding, exists := secondMap[ad.ID]
			require.True(s.T(), exists,
				"Объявление с ID %s из первого ответа не найдено во втором", ad.ID)

			assert.Equal(s.T(), ad.ID, corresponding.ID,
				"ID объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.SellerID, corresponding.SellerID,
				"SellerID объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.Name, corresponding.Name,
				"Name объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.Price, corresponding.Price,
				"Price объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.Statistics.Likes, corresponding.Statistics.Likes,
				"Statistics.Likes объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.Statistics.ViewCount, corresponding.Statistics.ViewCount,
				"Statistics.ViewCount объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.Statistics.Contacts, corresponding.Statistics.Contacts,
				"Statistics.Contacts объявления %d должен совпадать", i+1)
			assert.Equal(s.T(), ad.CreatedAt, corresponding.CreatedAt,
				"CreatedAt объявления %d должен совпадать", i+1)
		}
	})
}

// Нефункциональный: успешный ответ должен содержать заголовок Content-Type: application/json
func (s *TestSuite) TestGetAdvertisementsBySellerId_ResponseContentType() {
	sellerID := utils.RandomInt(111111, 999999)

	s.Run("Получение объявлений с дробным sellerID", func() {
		resp, _ := getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(
			s.T(),
			http.StatusOK,
			sellerID,
		)

		require.NotNil(s.T(), resp, "Response не должен быть nil")

		contentType := resp.Header.Get("Content-Type")

		assert.Equal(s.T(), "application/json", contentType,
			"Content-Type должен быть 'application/json', получено: %s", contentType)

		assert.NotEmpty(s.T(), contentType,
			"Content-Type заголовок не должен быть пустым")
	})
}

// Нефункциональный: время ответа при получении всех объявлений по sellerID не должно превышать 2000ms
func (s *TestSuite) TestGetAdvertisementsBySellerId_ResponseTime() {
	sellerID := utils.RandomInt(111111, 999999)
	var elapsed time.Duration

	s.Run("Создание трёх объявлений с одинаковым sellerID", func() {
		s.createAdsForSeller(sellerID, 3)
	})

	s.Run("Получение объявлений по sellerID с замером времени ответа", func() {
		start := time.Now()
		getAdvertisementsBySellerId.GetAdvertisementsBySellerIdInt(s.T(), http.StatusOK, sellerID)
		elapsed = time.Since(start)

		s.T().Logf("Время ответа: %s", elapsed)
	})

	s.Run("Проверка что время ответа не превышает 2000ms", func() {
		assert.LessOrEqual(s.T(), elapsed.Milliseconds(), int64(2000),
			"Время ответа %dms превышает допустимый порог в 2000ms", elapsed.Milliseconds())
	})
}

package createAdvertisement

import (
	"api-tests-template/internal/helpers/checkDeletedAdvertisementHelper"
	"api-tests-template/internal/helpers/createAdvertisementHelpers"
	"api-tests-template/internal/helpers/parseHelpers"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/createAdvertisement"
	"api-tests-template/internal/managers/deleteAdvertisement"
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

// parseAndRegisterID парсит тело ответа сервера, извлекает ID объявления из поля status
// и регистрирует его в списке на удаление после теста.
func (s *TestSuite) parseAndRegisterID(responseBody string) string {
	var response models.CreateAdvertisementResponseServer
	err := json.Unmarshal([]byte(responseBody), &response)
	require.NoError(s.T(), err, "Ошибка парсинга ответа сервера")

	id := parseHelpers.ParseIDFromStatus(response.Status)
	require.NotEmpty(s.T(), id, "ID объявления не должен быть пустым")

	s.createdAds = append(s.createdAds, id)
	return id
}

// assertBadResponse парсит тело ответа как BadResponse, проверяет HTTP-статус и сообщение об ошибке.
func (s *TestSuite) assertBadResponse(responseBody string, expectedMessage string) {
	var badRequest models.BadResponse
	err := json.Unmarshal([]byte(responseBody), &badRequest)
	require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

	assert.Equal(s.T(), "400", badRequest.Status,
		"Статус должен быть '400', получено: %s", badRequest.Status)

	assert.Equal(s.T(), expectedMessage, badRequest.Result.Message,
		"Некорректное сообщение об ошибке")

	assert.NotNil(s.T(), badRequest.Result.Messages, "Поле messages не должно быть nil")
}

// Позитивный: создание объявления со всеми полями и проверка возвращаемого ответа
func (s *TestSuite) TestPostAdvertisementSuccess() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()

	var responseBody string
	s.Run("Создание объявления со всеми полями", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(
			s.T(),
			http.StatusOK,
			*request,
		)
	})

	s.Run("Ответ содержит непустой id объявления", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err, "Тело ответа должно десериализоваться в CreateAdvertisementResponse")
		require.NotEmpty(s.T(), resp.ID, "Поле id не должно быть пустым")
	})

	s.Run("Ответ содержит корректный sellerId", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err)
		require.Equal(s.T(), request.SellerID, resp.SellerID, "sellerId должен совпадать с переданным в запросе")
	})

	s.Run("Ответ содержит корректное name", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err)
		require.Equal(s.T(), request.Name, resp.Name, "name должен совпадать с переданным в запросе")
	})

	s.Run("Ответ содержит корректный price", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err)
		require.Equal(s.T(), request.Price, resp.Price, "price должен совпадать с переданным в запросе")
	})

	s.Run("Ответ содержит корректную статистику", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), request.Statistics.Likes, resp.Statistics.Likes, "statistics.likes должен совпадать")
		assert.Equal(s.T(), request.Statistics.ViewCount, resp.Statistics.ViewCount, "statistics.viewCount должен совпадать")
		assert.Equal(s.T(), request.Statistics.Contacts, resp.Statistics.Contacts, "statistics.contacts должен совпадать")
	})

	s.Run("Ответ содержит непустой createdAt", func() {
		var resp models.CreateAdvertisementResponseDocumentation
		err := json.Unmarshal([]byte(responseBody), &resp)
		require.NoError(s.T(), err)
		require.NotEmpty(s.T(), resp.CreatedAt, "Поле createdAt не должно быть пустым")
	})
}

// Позитивный: два объявления с одинаковым sellerID должны получить разные уникальные ID
func (s *TestSuite) TestCreateAdvertisement_SameSellerID_UniqueIDs() {
	sellerID := utils.RandomInt(111111, 999999)

	requestFirst := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	requestFirst.SellerID = sellerID

	requestSecond := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	requestSecond.SellerID = sellerID

	var responseBodyFirstStr string
	var responseBodySecondStr string

	s.Run("Создание двух объявлений с одинаковым sellerID", func() {
		_, responseBodyFirstStr = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *requestFirst)
		_, responseBodySecondStr = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *requestSecond)
	})

	s.Run("Проверка что оба объявления получили разные уникальные ID", func() {
		var responseFirst, responseSecond models.CreateAdvertisementResponseServer

		err := json.Unmarshal([]byte(responseBodyFirstStr), &responseFirst)
		require.NoError(s.T(), err, "Ошибка парсинга первого ответа")

		err = json.Unmarshal([]byte(responseBodySecondStr), &responseSecond)
		require.NoError(s.T(), err, "Ошибка парсинга второго ответа")

		id1 := parseHelpers.ParseIDFromStatus(responseFirst.Status)
		id2 := parseHelpers.ParseIDFromStatus(responseSecond.Status)

		assert.NotEqual(s.T(), id1, id2,
			"Объявления с одинаковым sellerID должны иметь разные ID. Получены ID: %s и %s", id1, id2)

		s.createdAds = append(s.createdAds, id1)
		s.createdAds = append(s.createdAds, id2)
	})
}

// Позитивный: создание объявления с price = 0
func (s *TestSuite) TestCreateAdvertisementWithZeroPrice() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Price = 0

	var responseBody string

	s.Run("Создание объявления с ценой 0", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что объявление успешно создано и получило ID", func() {
		id := s.parseAndRegisterID(responseBody)

		var responseMap map[string]interface{}
		err := json.Unmarshal([]byte(responseBody), &responseMap)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		status, ok := responseMap["status"].(string)
		require.True(s.T(), ok, "Поле status не найдено в ответе")

		assert.Contains(s.T(), status, "Сохранили объявление - ",
			"Статус должен содержать 'Сохранили объявление - ', получено: %s", status)

		assert.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})
}

// Позитивный: два запроса с полностью одинаковыми параметрами должны создать разные объявления
func (s *TestSuite) TestCreateAdvertisement_IdenticalData_UniqueIDs() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	var responseBodyFirstStr string
	var responseBodySecondStr string

	s.Run("Создание двух объявлений с полностью одинаковыми параметрами", func() {
		_, responseBodyFirstStr = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
		_, responseBodySecondStr = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что оба запроса создали объявления с разными уникальными ID", func() {
		var responseFirst, responseSecond models.CreateAdvertisementResponseServer

		err := json.Unmarshal([]byte(responseBodyFirstStr), &responseFirst)
		require.NoError(s.T(), err, "Ошибка парсинга первого ответа")

		err = json.Unmarshal([]byte(responseBodySecondStr), &responseSecond)
		require.NoError(s.T(), err, "Ошибка парсинга второго ответа")

		id1 := parseHelpers.ParseIDFromStatus(responseFirst.Status)
		id2 := parseHelpers.ParseIDFromStatus(responseSecond.Status)

		assert.NotEqual(s.T(), id1, id2,
			"Объявления с одинаковыми параметрами должны иметь разные ID. Получены ID: %s и %s", id1, id2)

		s.createdAds = append(s.createdAds, id1)
		s.createdAds = append(s.createdAds, id2)
	})
}

// Позитивный: statistics со всеми нулевыми значениями является валидной и должна приниматься сервером
func (s *TestSuite) TestCreateAdvertisementWithZeroStatistics() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Statistics = models.Statistics{Likes: 0, ViewCount: 0, Contacts: 0}

	var responseBody string

	s.Run("Создание объявления с нулевой статистикой", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что объявление успешно создано и получило ID", func() {
		id := s.parseAndRegisterID(responseBody)

		var responseMap map[string]interface{}
		err := json.Unmarshal([]byte(responseBody), &responseMap)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		status, ok := responseMap["status"].(string)
		require.True(s.T(), ok, "Поле status не найдено в ответе")

		assert.Contains(s.T(), status, "Сохранили объявление - ",
			"Статус должен содержать 'Сохранили объявление - ', получено: %s", status)

		assert.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})
}

// Негативный: отсутствует обязательное поле name
func (s *TestSuite) TestCreateAdvertisement_MissingName_BadRequest() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Name = ""

	var responseBody string

	s.Run("Создание объявления с пустым полем name", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusBadRequest, *request)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(responseBody, "поле name обязательно")
	})
}

// Негативный: отсутствует обязательное поле sellerID
func (s *TestSuite) TestCreateAdvertisement_MissingSellerId_BadRequest() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.SellerID = 0

	var responseBody string

	s.Run("Создание объявления без sellerID (значение 0)", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusBadRequest, *request)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(responseBody, "поле sellerID обязательно")
	})
}

// Негативный: поле price содержит отрицательное значение
func (s *TestSuite) TestCreateAdvertisement_NegativePrice_BadRequest() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Price = -1

	var responseBody string

	s.Run("Создание объявления с отрицательным price", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusBadRequest, *request)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(responseBody, "поле price должно быть положительным")
	})
}

// Негативный: поле sellerID содержит отрицательное значение
func (s *TestSuite) TestCreateAdvertisement_NegativeSellerId_BadRequest() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.SellerID = -1

	var responseBody string

	s.Run("Создание объявления с отрицательным sellerID", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusBadRequest, *request)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(responseBody, "поле sellerID должно быть положительным")
	})
}

// Негативный: поле statistics содержит отрицательные значения
func (s *TestSuite) TestCreateAdvertisement_NegativeStatistics_BadRequest() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Statistics = models.Statistics{Likes: -1, ViewCount: -1, Contacts: -1}

	var responseBody string

	s.Run("Создание объявления с отрицательными значениями в statistics", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusBadRequest, *request)
	})

	s.Run("Проверка ответа сервера с ошибкой 400 и корректным сообщением", func() {
		s.assertBadResponse(responseBody, "поле Likes должно быть положительным")
	})
}

// Корнер-кейс: поле name содержит очень длинную строку
func (s *TestSuite) TestCreateAdvertisement_VeryLongName() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Name = utils.RandomString(10000)

	var responseBody string

	s.Run("Создание объявления с name длиной 10 000 символов", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что объявление успешно создано и получило ID", func() {
		id := s.parseAndRegisterID(responseBody)

		var responseMap map[string]interface{}
		err := json.Unmarshal([]byte(responseBody), &responseMap)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		status, ok := responseMap["status"].(string)
		require.True(s.T(), ok, "Поле status не найдено в ответе")

		assert.Contains(s.T(), status, "Сохранили объявление - ",
			"Статус должен содержать 'Сохранили объявление - ', получено: %s", status)

		assert.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})
}

// Корнер-кейс: поле name содержит спецсимволы и XSS-строку — данные должны сохраниться без изменений
func (s *TestSuite) TestCreateAdvertisement_SpecialCharsInName() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Name = "<script>alert('xss')</script>"

	var responseBody string

	s.Run("Создание объявления с name, содержащим спецсимволы", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что объявление успешно создано и получило ID", func() {
		id := s.parseAndRegisterID(responseBody)

		var responseMap map[string]interface{}
		err := json.Unmarshal([]byte(responseBody), &responseMap)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		status, ok := responseMap["status"].(string)
		require.True(s.T(), ok, "Поле status не найдено в ответе")

		assert.Contains(s.T(), status, "Сохранили объявление - ",
			"Статус должен содержать 'Сохранили объявление - ', получено: %s", status)

		assert.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})
}

// Корнер-кейс: поле price содержит максимально большое значение
func (s *TestSuite) TestCreateAdvertisement_MaxPrice_Success() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Price = 999999999999999

	var responseBody string

	s.Run("Создание объявления с максимально большим price", func() {
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
	})

	s.Run("Проверка что объявление успешно создано и получило ID", func() {
		id := s.parseAndRegisterID(responseBody)

		var responseMap map[string]interface{}
		err := json.Unmarshal([]byte(responseBody), &responseMap)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		status, ok := responseMap["status"].(string)
		require.True(s.T(), ok, "Поле status не найдено в ответе")

		assert.Contains(s.T(), status, "Сохранили объявление - ",
			"Статус должен содержать 'Сохранили объявление - ', получено: %s", status)

		assert.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})
}

// Нефункциональный: успешный ответ должен содержать заголовок Content-Type: application/json
func (s *TestSuite) TestCreateAdvertisement_ResponseContentType() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()

	var resp *http.Response
	var body string

	s.Run("Создание объявления", func() {
		resp, body = createAdvertisement.CreateAdvertisement(
			s.T(),
			http.StatusOK,
			*request,
		)
	})

	s.Run("Проверка заголовка Content-Type: application/json", func() {
		id := s.parseAndRegisterID(body)
		require.NotEmpty(s.T(), id, "ID не должен быть пустым")

		require.NotNil(s.T(), resp, "Response не должен быть nil")

		contentType := resp.Header.Get("Content-Type")

		assert.Equal(s.T(), "application/json", contentType,
			"Content-Type должен быть 'application/json', получено: %s", contentType)

		assert.NotEmpty(s.T(), contentType,
			"Content-Type заголовок не должен быть пустым")
	})
}

// Нефункциональный: время ответа при создании объявления не должно превышать 2000ms
func (s *TestSuite) TestCreateAdvertisement_ResponseTime() {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()

	var elapsed time.Duration
	var responseBody string

	s.Run("Создание объявления с замером времени ответа", func() {
		start := time.Now()
		_, responseBody = createAdvertisement.CreateAdvertisement(s.T(), http.StatusOK, *request)
		elapsed = time.Since(start)

		s.T().Logf("Время ответа: %s", elapsed)
	})

	s.Run("Регистрация созданного объявления на удаление", func() {
		id := s.parseAndRegisterID(responseBody)
		require.NotEmpty(s.T(), id, "ID не должен быть пустым")
	})

	s.Run("Проверка что время ответа не превышает 2000ms", func() {
		assert.LessOrEqual(s.T(), elapsed.Milliseconds(), int64(2000),
			"Время ответа %dms превышает допустимый порог в 2000ms", elapsed.Milliseconds())
	})
}

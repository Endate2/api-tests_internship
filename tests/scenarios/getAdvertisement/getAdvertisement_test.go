package getAdvertisement

import (
	"api-tests-template/internal/helpers/checkDeletedAdvertisementHelper"
	"api-tests-template/internal/helpers/createAdvertisementHelpers"
	"api-tests-template/internal/helpers/parseHelpers"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/createAdvertisement"
	"api-tests-template/internal/managers/deleteAdvertisement"
	"api-tests-template/internal/managers/getAdvertisement"
	"api-tests-template/internal/utils"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	base "api-tests-template/tests"

	"github.com/google/uuid"
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

// createAndRegisterAd создаёт объявление, извлекает его ID из ответа и регистрирует на удаление.
// Возвращает ID созданного объявления и исходный запрос.
func (s *TestSuite) createAndRegisterAd() (string, *models.CreateAdvertisementRequest) {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()

	_, responseBody := createAdvertisement.CreateAdvertisement(
		s.T(),
		http.StatusOK,
		*request,
	)

	var response models.CreateAdvertisementResponseServer
	err := json.Unmarshal([]byte(responseBody), &response)
	require.NoError(s.T(), err, "Ошибка парсинга ответа при создании объявления")

	id := parseHelpers.ParseIDFromStatus(response.Status)
	require.NotEmpty(s.T(), id, "ID созданного объявления не должен быть пустым")

	s.createdAds = append(s.createdAds, id)
	return id, request
}

// assertBadResponse парсит тело ответа как BadResponse и проверяет код ошибки и сообщение.
func (s *TestSuite) assertBadResponse(responseBody string, expectedStatus string, expectedMessageContains string) {
	var badRequest models.BadResponse
	err := json.Unmarshal([]byte(responseBody), &badRequest)
	require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

	assert.Equal(s.T(), expectedStatus, badRequest.Status,
		"Статус должен быть '%s', получено: %s", expectedStatus, badRequest.Status)

	assert.Contains(s.T(), badRequest.Result.Message, expectedMessageContains,
		"Сообщение об ошибке должно содержать '%s'", expectedMessageContains)

	assert.NotNil(s.T(), badRequest.Result.Messages, "Поле messages не должно быть nil")
}

// Позитивный: после создания объявления GET /item/:id должен вернуть те же данные, что были переданы при создании
func (s *TestSuite) TestGetAdvertisementSuccess() {
	var createdAdID string
	var request *models.CreateAdvertisementRequest

	s.Run("Создание объявления", func() {
		createdAdID, request = s.createAndRegisterAd()
	})

	var ad models.GetAdvertisementResponse
	s.Run("Получение объявления по ID", func() {
		_, getResponseBody := getAdvertisement.GetAdvertisement(s.T(), http.StatusOK, createdAdID)

		var advertisements models.GetAdvertisementResponseList
		err := json.Unmarshal([]byte(getResponseBody), &advertisements)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		require.NotEmpty(s.T(), advertisements, "Ответ не должен быть пустым массивом")

		ad = advertisements[0]
	})

	s.Run("Проверка что все поля совпадают с данными, переданными при создании", func() {
		assert.Equal(s.T(), createdAdID, ad.ID, "ID должен совпадать")
		assert.Equal(s.T(), request.SellerID, ad.SellerID, "SellerID должен совпадать")
		assert.Equal(s.T(), request.Name, ad.Name, "Name должен совпадать")
		assert.Equal(s.T(), request.Price, ad.Price, "Price должен совпадать")
		assert.Equal(s.T(), request.Statistics.Likes, ad.Statistics.Likes, "Statistics.Likes должен совпадать")
		assert.Equal(s.T(), request.Statistics.ViewCount, ad.Statistics.ViewCount, "Statistics.ViewCount должен совпадать")
		assert.Equal(s.T(), request.Statistics.Contacts, ad.Statistics.Contacts, "Statistics.Contacts должен совпадать")
		assert.NotEmpty(s.T(), ad.CreatedAt, "CreatedAt не должен быть пустым")
	})
}

// Негативный: запрос с невалидным ID должен вернуть 400
func (s *TestSuite) TestGetAdvertisement_InvalidUUID_BadRequest() {
	var getResponseBody string

	s.Run("Получение объявления с невалидным ID (не UUID)", func() {
		invalidID := utils.RandomString(50)
		_, getResponseBody = getAdvertisement.GetAdvertisement(s.T(), http.StatusBadRequest, invalidID)
	})

	s.Run("Проверка ответа сервера с ошибкой 400", func() {
		s.assertBadResponse(getResponseBody, "400", "ID айтема не UUID")
	})
}

// Негативный: запрос с валидным UUID, которого нет в базе, должен вернуть 404
func (s *TestSuite) TestGetAdvertisement_NonExistentUUID_NotFound() {
	nonExistentUUID := uuid.New().String()

	var getResponseBody string

	s.Run("Получение объявления по несуществующему UUID", func() {
		_, getResponseBody = getAdvertisement.GetAdvertisement(s.T(), http.StatusNotFound, nonExistentUUID)
	})

	s.Run("Проверка ответа сервера с ошибкой 404", func() {
		var badRequest models.BadResponse
		err := json.Unmarshal([]byte(getResponseBody), &badRequest)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		assert.Equal(s.T(), "404", badRequest.Status,
			"Статус должен быть '404', получено: %s", badRequest.Status)

		assert.Contains(s.T(), badRequest.Result.Message, "item "+nonExistentUUID+" not found",
			"Сообщение должно указывать, что объявление не найдено")

		assert.Empty(s.T(), badRequest.Result.Messages, "Поле messages должно быть пустым")
	})
}

// Негативный: запрос с пустым ID должен вернуть 404
func (s *TestSuite) TestGetAdvertisement_EmptyID_NotFound() {
	var getResponseBody string

	s.Run("Получение объявления с пустым ID", func() {
		_, getResponseBody = getAdvertisement.GetAdvertisement(s.T(), http.StatusNotFound, "")
	})

	s.Run("Проверка ответа сервера с ошибкой 404", func() {
		var badRequest models.BadResponse
		err := json.Unmarshal([]byte(getResponseBody), &badRequest)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		assert.Equal(s.T(), "404", badRequest.Status,
			"Статус должен быть '404', получено: %s", badRequest.Status)

		assert.Contains(s.T(), badRequest.Result.Message, "item not found",
			"Сообщение должно указывать, что объявление не найдено")

		assert.Empty(s.T(), badRequest.Result.Messages, "Поле messages должно быть пустым")
	})
}

// Негативный: запрос с очень длинной строкой вместо UUID должен вернуть 400
func (s *TestSuite) TestGetAdvertisement_VeryLongID_BadRequest() {
	var getResponseBody string

	s.Run("Получение объявления с ID длиной 400 символов", func() {
		invalidID := utils.RandomString(400)
		_, getResponseBody = getAdvertisement.GetAdvertisement(s.T(), http.StatusBadRequest, invalidID)
	})

	s.Run("Проверка ответа сервера с ошибкой 400", func() {
		s.assertBadResponse(getResponseBody, "400", "ID айтема не UUID")
	})
}

// Корнер-кейс: повторные запросы с одним и тем же ID должны возвращать идентичные данные (идемпотентность GET)
func (s *TestSuite) TestGetAdvertisement_Idempotency() {
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	var ad1, ad2 models.GetAdvertisementResponse
	s.Run("Два последовательных GET-запроса по одному ID", func() {
		_, getResponseBodyFirst := getAdvertisement.GetAdvertisement(s.T(), http.StatusOK, createdAdID)
		_, getResponseBodySecond := getAdvertisement.GetAdvertisement(s.T(), http.StatusOK, createdAdID)

		var advertisementsFirst, advertisementsSecond models.GetAdvertisementResponseList
		errFirst := json.Unmarshal([]byte(getResponseBodyFirst), &advertisementsFirst)
		errSecond := json.Unmarshal([]byte(getResponseBodySecond), &advertisementsSecond)
		require.NoError(s.T(), errFirst, "Ошибка парсинга первого ответа")
		require.NoError(s.T(), errSecond, "Ошибка парсинга второго ответа")

		require.NotEmpty(s.T(), advertisementsFirst, "Первый ответ не должен быть пустым массивом")
		require.NotEmpty(s.T(), advertisementsSecond, "Второй ответ не должен быть пустым массивом")

		ad1 = advertisementsFirst[0]
		ad2 = advertisementsSecond[0]
	})

	s.Run("Проверка что оба ответа содержат одинаковые данные", func() {
		assert.Equal(s.T(), ad1.ID, ad2.ID, "ID должен совпадать")
		assert.Equal(s.T(), ad1.SellerID, ad2.SellerID, "SellerID должен совпадать")
		assert.Equal(s.T(), ad1.Name, ad2.Name, "Name должен совпадать")
		assert.Equal(s.T(), ad1.Price, ad2.Price, "Price должен совпадать")
		assert.Equal(s.T(), ad1.Statistics.Likes, ad2.Statistics.Likes, "Statistics.Likes должен совпадать")
		assert.Equal(s.T(), ad1.Statistics.ViewCount, ad2.Statistics.ViewCount, "Statistics.ViewCount должен совпадать")
		assert.Equal(s.T(), ad1.Statistics.Contacts, ad2.Statistics.Contacts, "Statistics.Contacts должен совпадать")
		assert.NotEmpty(s.T(), ad1.CreatedAt, "CreatedAt первого ответа не должен быть пустым")
		assert.NotEmpty(s.T(), ad2.CreatedAt, "CreatedAt второго ответа не должен быть пустым")
	})
}

// Корнер-кейс: ID содержит спецсимволы — должен быть отклонён как невалидный UUID
func (s *TestSuite) TestGetAdvertisement_SpecialSymbolsInID_BadRequest() {
	var getResponseBody string
	invalidID := "<script>alert('xss')</script>"

	s.Run("Получение объявления с ID, содержащим спецсимволы", func() {
		_, getResponseBody = getAdvertisement.GetAdvertisement(s.T(), http.StatusBadRequest, invalidID)
	})

	s.Run("Проверка ответа сервера с ошибкой 400", func() {
		s.assertBadResponse(getResponseBody, "400", "ID айтема не UUID")
	})
}

// Нефункциональный: успешный ответ должен содержать заголовок Content-Type: application/json
func (s *TestSuite) TestGetAdvertisement_ResponseContentType() {
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	s.Run("Два последовательных GET-запроса по одному ID", func() {
		resp, getResponseBodyFirst := getAdvertisement.GetAdvertisement(s.T(), http.StatusOK, createdAdID)

		var advertisementsFirst models.GetAdvertisementResponseList
		errFirst := json.Unmarshal([]byte(getResponseBodyFirst), &advertisementsFirst)
		require.NoError(s.T(), errFirst, "Ошибка парсинга первого ответа")

		require.NotNil(s.T(), resp, "Response не должен быть nil")

		contentType := resp.Header.Get("Content-Type")

		assert.Equal(s.T(), "application/json", contentType,
			"Content-Type должен быть 'application/json', получено: %s", contentType)

		assert.NotEmpty(s.T(), contentType,
			"Content-Type заголовок не должен быть пустым")
	})
}

// Нефункциональный: время ответа при получении объявления по ID не должно превышать 2000ms
func (s *TestSuite) TestGetAdvertisement_ResponseTime() {
	var elapsed time.Duration
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	s.Run("Получение объявления по ID с замером времени ответа", func() {
		start := time.Now()
		getAdvertisement.GetAdvertisement(s.T(), http.StatusOK, createdAdID)
		elapsed = time.Since(start)

		s.T().Logf("Время ответа: %s", elapsed)
	})

	s.Run("Проверка что время ответа не превышает 2000ms", func() {
		assert.LessOrEqual(s.T(), elapsed.Milliseconds(), int64(2000),
			"Время ответа %dms превышает допустимый порог в 2000ms", elapsed.Milliseconds())
	})
}

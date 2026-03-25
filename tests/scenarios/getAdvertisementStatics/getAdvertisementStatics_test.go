package getAdvertisementStatics

import (
	"api-tests-template/internal/helpers/checkDeletedAdvertisementHelper"
	"api-tests-template/internal/helpers/createAdvertisementHelpers"
	"api-tests-template/internal/helpers/parseHelpers"
	"api-tests-template/internal/managers/auth/models"
	"api-tests-template/internal/managers/createAdvertisement"
	"api-tests-template/internal/managers/deleteAdvertisement"
	"api-tests-template/internal/managers/getAdvertisementStatics"
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

// createAndRegisterAdWithStatistics создаёт объявление с заданной статистикой,
// регистрирует его на удаление и возвращает ID и исходный запрос.
func (s *TestSuite) createAndRegisterAdWithStatistics(stats models.Statistics) (string, *models.CreateAdvertisementRequest) {
	request := createAdvertisementHelpers.CreateTestAdvertisementHelper()
	request.Statistics = stats

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

// Позитивный: статистика созданного объявления должна совпадать с переданной при создании
func (s *TestSuite) TestGetAdvertisementStatics_Success() {
	stats := models.Statistics{
		Likes:     utils.RandomInt(1, 100),
		ViewCount: utils.RandomInt(1, 100),
		Contacts:  utils.RandomInt(1, 100),
	}

	var createdAdID string
	var request *models.CreateAdvertisementRequest

	s.Run("Создание объявления с заданной статистикой", func() {
		createdAdID, request = s.createAndRegisterAdWithStatistics(stats)
	})

	var staticsResponseBody string
	s.Run("Получение статистики по ID объявления", func() {
		_, staticsResponseBody = getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)
		s.T().Logf("Response body: %s", staticsResponseBody)
	})

	s.Run("Проверка что статистика возвращается в виде массива с объектом", func() {
		var statisticsList []models.Statistics
		err := json.Unmarshal([]byte(staticsResponseBody), &statisticsList)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")
		require.NotEmpty(s.T(), statisticsList, "Ответ не должен быть пустым массивом")
	})

	s.Run("Проверка что значения статистики совпадают с переданными при создании", func() {
		var statisticsList []models.Statistics
		err := json.Unmarshal([]byte(staticsResponseBody), &statisticsList)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")

		stat := statisticsList[0]
		assert.Equal(s.T(), request.Statistics.Likes, stat.Likes,
			"Likes должен совпадать: ожидалось %d, получено %d", request.Statistics.Likes, stat.Likes)
		assert.Equal(s.T(), request.Statistics.ViewCount, stat.ViewCount,
			"ViewCount должен совпадать: ожидалось %d, получено %d", request.Statistics.ViewCount, stat.ViewCount)
		assert.Equal(s.T(), request.Statistics.Contacts, stat.Contacts,
			"Contacts должен совпадать: ожидалось %d, получено %d", request.Statistics.Contacts, stat.Contacts)
	})
}

// Позитивный: поля likes, viewCount, contacts должны быть целыми числами >= 0
func (s *TestSuite) TestGetAdvertisementStatics_FieldsAreNonNegativeIntegers() {
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	var statisticsList []models.Statistics
	s.Run("Получение статистики по ID объявления", func() {
		_, responseBody := getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)

		err := json.Unmarshal([]byte(responseBody), &statisticsList)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")
		require.NotEmpty(s.T(), statisticsList, "Ответ не должен быть пустым массивом")
	})

	s.Run("Проверка что likes, viewCount, contacts — целые числа не меньше 0", func() {
		stat := statisticsList[0]
		assert.GreaterOrEqual(s.T(), stat.Likes, 0,
			"Likes должен быть >= 0, получено: %d", stat.Likes)
		assert.GreaterOrEqual(s.T(), stat.ViewCount, 0,
			"ViewCount должен быть >= 0, получено: %d", stat.ViewCount)
		assert.GreaterOrEqual(s.T(), stat.Contacts, 0,
			"Contacts должен быть >= 0, получено: %d", stat.Contacts)
	})
}

// Позитивный статистика с нулевыми значениями должна корректно возвращаться
func (s *TestSuite) TestGetAdvertisementStatics_ZeroStatistics() {
	stats := models.Statistics{
		Likes:     0,
		ViewCount: 0,
		Contacts:  0,
	}

	var createdAdID string

	s.Run("Создание объявления с нулевой статистикой", func() {
		createdAdID, _ = s.createAndRegisterAdWithStatistics(stats)
	})

	var statisticsList []models.Statistics
	s.Run("Получение статистики по ID объявления", func() {
		_, responseBody := getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)

		err := json.Unmarshal([]byte(responseBody), &statisticsList)
		require.NoError(s.T(), err, "Ошибка парсинга ответа")
		require.NotEmpty(s.T(), statisticsList, "Ответ не должен быть пустым массивом")
	})

	s.Run("Проверка что нулевые значения статистики корректно отображаются", func() {
		stat := statisticsList[0]
		assert.Equal(s.T(), 0, stat.Likes,
			"Likes должен быть 0, получено: %d", stat.Likes)
		assert.Equal(s.T(), 0, stat.ViewCount,
			"ViewCount должен быть 0, получено: %d", stat.ViewCount)
		assert.Equal(s.T(), 0, stat.Contacts,
			"Contacts должен быть 0, получено: %d", stat.Contacts)
	})
}

// Негативный: запрос статистики по несуществующему UUID должен вернуть 404
func (s *TestSuite) TestGetAdvertisementStatics_NonExistentUUID_NotFound() {
	nonExistentUUID := uuid.New().String()

	var responseBody string

	s.Run("Получение статистики по несуществующему UUID", func() {
		_, responseBody = getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusNotFound,
			nonExistentUUID,
		)
		s.T().Logf("Response body: %s", responseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 404", func() {
		var badRequest models.BadResponse
		err := json.Unmarshal([]byte(responseBody), &badRequest)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		assert.Equal(s.T(), "404", badRequest.Status,
			"Статус должен быть '404', получено: %s", badRequest.Status)

		assert.Contains(s.T(), badRequest.Result.Message, nonExistentUUID,
			"Сообщение об ошибке должно содержать ID: %s", nonExistentUUID)
	})
}

// Негативный: запрос статистики с невалидным ID (не UUID) должен вернуть 400
func (s *TestSuite) TestGetAdvertisementStatics_InvalidUUID_BadRequest() {
	var responseBody string

	s.Run("Получение статистики с невалидным ID (не UUID)", func() {
		invalidID := utils.RandomString(50)
		_, responseBody = getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusBadRequest,
			invalidID,
		)
		s.T().Logf("Response body: %s", responseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 400", func() {
		s.assertBadResponse(responseBody, "400", "ID айтема не UUID")
	})
}

// Негативный: запрос статистики с очень длинным ID (400 символов) должен вернуть 400
func (s *TestSuite) TestGetAdvertisementStatics_VeryLongID_BadRequest() {
	var responseBody string

	s.Run("Получение статистики с ID длиной 400 символов", func() {
		invalidID := utils.RandomString(400)
		_, responseBody = getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusBadRequest,
			invalidID,
		)
		s.T().Logf("Response body: %s", responseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 400", func() {
		s.assertBadResponse(responseBody, "400", "ID айтема не UUID")
	})
}

// Негативный: запрос статистики с пустым ID должен вернуть 404
func (s *TestSuite) TestGetAdvertisementStatics_EmptyID_NotFound() {
	var responseBody string

	s.Run("Получение статистики с пустым ID", func() {
		_, responseBody = getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusNotFound,
			"",
		)
		s.T().Logf("Response body: %s", responseBody)
	})

	s.Run("Проверка ответа сервера с ошибкой 404", func() {
		var badRequest models.NotFoundResponse
		err := json.Unmarshal([]byte(responseBody), &badRequest)
		require.NoError(s.T(), err, "Ошибка парсинга ответа - не соответствует документации")

		assert.Equal(s.T(), 404, badRequest.Code,
			"Статус должен быть '404', получено: %s", badRequest.Code)
	})
}

// Корнер-кейс: повторные запросы статистики по одному ID должны возвращать идентичные данные (идемпотентность)
func (s *TestSuite) TestGetAdvertisementStatics_Idempotency() {
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	var firstStats, secondStats []models.Statistics
	s.Run("Два последовательных GET-запроса статистики по одному ID", func() {
		_, firstBody := getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)
		_, secondBody := getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)

		errFirst := json.Unmarshal([]byte(firstBody), &firstStats)
		errSecond := json.Unmarshal([]byte(secondBody), &secondStats)
		require.NoError(s.T(), errFirst, "Ошибка парсинга первого ответа")
		require.NoError(s.T(), errSecond, "Ошибка парсинга второго ответа")

		require.NotEmpty(s.T(), firstStats, "Первый ответ не должен быть пустым массивом")
		require.NotEmpty(s.T(), secondStats, "Второй ответ не должен быть пустым массивом")
	})

	s.Run("Проверка что оба ответа содержат одинаковые данные", func() {
		assert.Equal(s.T(), firstStats[0].Likes, secondStats[0].Likes,
			"Likes должен совпадать в обоих ответах")
		assert.Equal(s.T(), firstStats[0].ViewCount, secondStats[0].ViewCount,
			"ViewCount должен совпадать в обоих ответах")
		assert.Equal(s.T(), firstStats[0].Contacts, secondStats[0].Contacts,
			"Contacts должен совпадать в обоих ответах")
	})
}

// Нефункциональный: время ответа при получении статистики не должно превышать 2000ms
func (s *TestSuite) TestGetAdvertisementStatics_ResponseTime() {
	var createdAdID string
	var elapsed time.Duration

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	s.Run("Получение статистики с замером времени ответа", func() {
		start := time.Now()
		getAdvertisementStatics.GetAdvertisementStatics(s.T(), http.StatusOK, createdAdID)
		elapsed = time.Since(start)

		s.T().Logf("Время ответа: %s", elapsed)
	})

	s.Run("Проверка что время ответа не превышает 2000ms", func() {
		assert.LessOrEqual(s.T(), elapsed.Milliseconds(), int64(2000),
			"Время ответа %dms превышает допустимый порог в 2000ms", elapsed.Milliseconds())
	})
}

// Нефункцилальный: ответ должен содержать Content-Type: application/json
func (s *TestSuite) TestGetAdvertisementStatics_ResponseContentType() {
	var createdAdID string

	s.Run("Создание объявления", func() {
		createdAdID, _ = s.createAndRegisterAd()
	})

	s.Run("Получение статистики и проверка заголовка Content-Type", func() {
		resp, _ := getAdvertisementStatics.GetAdvertisementStatics(
			s.T(),
			http.StatusOK,
			createdAdID,
		)

		require.NotNil(s.T(), resp, "Response не должен быть nil")

		contentType := resp.Header.Get("Content-Type")
		assert.Equal(s.T(), "application/json", contentType,
			"Content-Type должен быть 'application/json', получено: %s", contentType)
		assert.NotEmpty(s.T(), contentType, "Content-Type заголовок не должен быть пустым")
	})
}

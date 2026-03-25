package createAdvertisement

import (
	"api-tests-template/internal/client/http/advertisementPost"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"api-tests-template/internal/managers/auth/models"
)

func CreateAdvertisement(t *testing.T, expectedStatusCode int, adData models.CreateAdvertisementRequest) (*http.Response, string) {
	createAdvertisementResponse := advertisementPost.HttpPostAdvertisement(t, adData)
	body, errResponse := io.ReadAll(createAdvertisementResponse.Body)

	t.Logf("Status: %d, Body: %s", createAdvertisementResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, createAdvertisementResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	require.NoError(t, errResponse)

	return createAdvertisementResponse, string(body)
}

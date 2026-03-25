package getAdvertisement

import (
	"api-tests-template/internal/client/http/advertisementGet"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetAdvertisement(t *testing.T, expectedStatusCode int, id string) (*http.Response, string) {
	getAdvertisementResponse := advertisementGet.HttpGetAdvertisement(t, id)
	body, errResponse := io.ReadAll(getAdvertisementResponse.Body)

	t.Logf("Status: %d, Body: %s", getAdvertisementResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, getAdvertisementResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	require.NoError(t, errResponse)

	return getAdvertisementResponse, string(body)
}

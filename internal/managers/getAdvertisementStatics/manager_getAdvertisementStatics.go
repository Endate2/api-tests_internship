package getAdvertisementStatics

import (
	"api-tests-template/internal/client/http/advertisementStatics"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetAdvertisementStatics(t *testing.T, expectedStatusCode int, id string) (*http.Response, string) {
	getAdvertisementStaticsResponse := advertisementStatics.HttpGetAdvertisementStatics(t, id)
	body, errResponse := io.ReadAll(getAdvertisementStaticsResponse.Body)

	t.Logf("Status: %d, Body: %s", getAdvertisementStaticsResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, getAdvertisementStaticsResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	require.NoError(t, errResponse)

	return getAdvertisementStaticsResponse, string(body)
}

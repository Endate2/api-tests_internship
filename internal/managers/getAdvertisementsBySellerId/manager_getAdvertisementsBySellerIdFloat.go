package getAdvertisementsBySellerId

import (
	"api-tests-template/internal/client/http/advertisementSellerIdFloat"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetAdvertisementsBySellerIdFloat(t *testing.T, expectedStatusCode int, id float64) (*http.Response, string) {
	getAdvertisementsBySellerIdResponse := advertisementSellerIdFloat.HttpGetAdvertisementBySellerIdFloat(t, id)
	body, errResponse := io.ReadAll(getAdvertisementsBySellerIdResponse.Body)

	t.Logf("Status: %d, Body: %s", getAdvertisementsBySellerIdResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, getAdvertisementsBySellerIdResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	require.NoError(t, errResponse)

	return getAdvertisementsBySellerIdResponse, string(body)
}

package getAdvertisementsBySellerId

import (
	"api-tests-template/internal/client/http/advertisementSellerIdInt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetAdvertisementsBySellerIdInt(t *testing.T, expectedStatusCode int, id int) (*http.Response, string) {
	getAdvertisementsBySellerIdResponse := advertisementSellerIdInt.HttpGetAdvertisementBySellerIdInt(t, id)
	body, errResponse := io.ReadAll(getAdvertisementsBySellerIdResponse.Body)

	t.Logf("Status: %d, Body: %s", getAdvertisementsBySellerIdResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, getAdvertisementsBySellerIdResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	require.NoError(t, errResponse)

	return getAdvertisementsBySellerIdResponse, string(body)
}

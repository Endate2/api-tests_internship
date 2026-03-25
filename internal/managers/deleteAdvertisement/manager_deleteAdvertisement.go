package deleteAdvertisement

import (
	"api-tests-template/internal/client/http/advertisementDelete"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func DeleteAdvertisement(t *testing.T, expectedStatusCode int, id string) string {
	deleteAdvetrisementResponse := advertisementDelete.HttpDeleteAdvertisement(t, id)

	body, err := io.ReadAll(deleteAdvetrisementResponse.Body)
	require.NoError(t, err)

	t.Logf("Status: %d, Body: %s", deleteAdvetrisementResponse.StatusCode, string(body))

	require.Equalf(t, expectedStatusCode, deleteAdvetrisementResponse.StatusCode,
		"HTTP status code должен быть %d", expectedStatusCode)

	return string(body)
}

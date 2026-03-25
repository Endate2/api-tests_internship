package advertisementSellerIdFloat

import (
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func HttpGetAdvertisementBySellerIdFloat(t *testing.T, id float64) *http.Response {
	idStr := strconv.FormatFloat(id, 'f', -1, 64)
	getPath := fmt.Sprintf("/1/%s/item", idStr)

	return apiRunner.GetRunner().Create().Get(getPath).
		ContentType("application/json").
		Expect(t).
		End().Response
}

package advertisementSellerIdInt

import (
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

func HttpGetAdvertisementBySellerIdInt(t *testing.T, id int) *http.Response {
	idStr := strconv.FormatInt(int64(id), 10)
	getPath := fmt.Sprintf("/1/%s/item", idStr)

	return apiRunner.GetRunner().Create().Get(getPath).
		ContentType("application/json").
		Expect(t).
		End().Response
}

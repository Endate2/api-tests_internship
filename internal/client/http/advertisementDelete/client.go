package advertisementDelete

import (
	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"fmt"
	"net/http"
	"testing"
)

func HttpDeleteAdvertisement(t *testing.T, id string) *http.Response {
	deletePath := fmt.Sprintf("%s/%s", path.DeleteAdvertisement, id)

	return apiRunner.GetRunner().Create().Delete(deletePath).
		ContentType("application/json").
		Expect(t).
		End().Response
}

package advertisementGet

import (
	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"fmt"
	"net/http"
	"testing"
)

func HttpGetAdvertisement(t *testing.T, id string) *http.Response {
	getPath := fmt.Sprintf("%s/%s", path.Advertisement, id)

	return apiRunner.GetRunner().Create().Get(getPath).
		ContentType("application/json").
		Expect(t).
		End().Response
}

package advertisementStatics

import (
	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
	"fmt"
	"net/http"
	"testing"
)

func HttpGetAdvertisementStatics(t *testing.T, id string) *http.Response {
	getPath := fmt.Sprintf("%s/%s", path.StaticsAdvertisement, id)

	return apiRunner.GetRunner().Create().Get(getPath).
		ContentType("application/json").
		Expect(t).
		End().Response
}

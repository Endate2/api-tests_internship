package advertisementPost

import (
	"api-tests-template/internal/managers/auth/models"
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

func HttpPostAdvertisement(t *testing.T, adData models.CreateAdvertisementRequest) *http.Response {
	runner := apiRunner.GetRunner().Create().Post(path.Advertisement).
		JSON(adData)

	return runner.Expect(t).End().Response
}

package checkDeletedAdvertisementHelper

import (
	"api-tests-template/internal/client/http/advertisementGet"
	"net/http"
	"testing"
)

func VerifyAdvertisementDeleted(t *testing.T, id string) bool {
	getResponse := advertisementGet.HttpGetAdvertisement(t, id)

	if getResponse.StatusCode == http.StatusNotFound {
		t.Logf("Объявление с ID %s успешно удалено (возвращает 404)", id)
		return true
	}

	if getResponse.StatusCode == http.StatusOK {
		t.Logf("Объявление с ID %s всё ещё существует", id)
		return false
	}

	t.Logf("Неожиданный статус код %d при проверке объявления %s", getResponse.StatusCode, id)
	return false
}

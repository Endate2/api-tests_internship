package api_runner

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/steinfletcher/apitest"
)

// Этот файл представляет собой утилиту-обёртку для библиотеки apitest
// предназначенную для упрощения написания и выполнения HTTP-тестов с авторизацией и сетевыми настройками.
const (
	defaultUserAgent = "Chrome/143.0.0.0"
)

type ApiTest struct {
	userAgent string
	host      string
}

func New() *ApiTest {
	return &ApiTest{
		userAgent: defaultUserAgent,
		host:      os.Getenv("API_URL"),
	}
}

func GetRunner() *ApiTest {
	return New()
}

func (at *ApiTest) Auth() *ApiTest {
	return at
}

func (at *ApiTest) Create() *apitest.APITest {
	apitestNew := apitest.New().EnableNetworking()
	if os.Getenv("DEBUG") == "true" {
		apitestNew = apitestNew.Debug()
	}
	return apitestNew.
		Intercept(func(request *http.Request) {
			request.Header.Set("User-Agent", at.userAgent)

			_ = MergeServiceUrls(at.host, request.URL)
		})
}

func MergeServiceUrls(host string, r *url.URL) error {
	urlParsed, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("host cannot be parsed: %s", err.Error())
	}

	if urlParsed.Path != "" {
		r.Path = path.Join(urlParsed.Path, r.Path)
	}

	r.Scheme = urlParsed.Scheme
	r.Host = urlParsed.Host
	return nil
}

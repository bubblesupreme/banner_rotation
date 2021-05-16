package integration_tests

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func addBanner(url string, descr string) (int, error) {
	request := struct {
		BannerURL   string `json:"url"`
		BannerDescr string `json:"description"`
	}{
		BannerURL: url,
		BannerDescr: descr,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	r := bytes.NewReader(jsonRequest)

	resp, err := http.Post("http://127.0.0.1:8088/banner", "application/json", r)

	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

func TestThompsonAllCold(t *testing.T) {
	status, err := addBanner("https://mybanner.com/banner1", "banner1")
	assert.NoError(t, err)
	assert.Equal(t, 200, status)
}

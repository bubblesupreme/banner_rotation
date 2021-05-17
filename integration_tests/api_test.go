package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/bubblesupreme/banner_rotation/internal/repository"
	"github.com/stretchr/testify/assert"
)

func addBanner(url string, descr string) (*repository.Banner, error) {
	request := struct {
		BannerURL   string `json:"url"`
		BannerDescr string `json:"description"`
	}{
		BannerURL:   url,
		BannerDescr: descr,
	}

	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(jsonRequest)

	resp, err := http.Post("http://127.0.0.1:8088/banner", "application/json", r) //nolint:noctx
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to add new banner")
	}
	defer resp.Body.Close()

	b := repository.Banner{}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &b); err != nil {
		return nil, err
	}

	return &b, nil
}

func addSlot() (*repository.Slot, error) {
	resp, err := http.Post("http://127.0.0.1:8088/slot", "application/json", nil) //nolint:noctx
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to add new slot")
	}
	defer resp.Body.Close()
	s := repository.Slot{}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

func addGroup(desc string) (*repository.Group, error) {
	reqData := struct {
		GroupDescr string `json:"description"`
	}{
		GroupDescr: desc,
	}

	req, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://127.0.0.1:8088/group", "application/json", bytes.NewReader(req)) //nolint:noctx
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to add new group")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	g := repository.Group{}
	if err := json.Unmarshal(body, &g); err != nil {
		return nil, err
	}

	return &g, nil
}

func addRelation(slotID, bannerID int) error {
	reqData := struct {
		SlotID   int `json:"slot"`
		BannerID int `json:"banner"`
	}{
		SlotID:   slotID,
		BannerID: bannerID,
	}
	req, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://127.0.0.1:8088/relation", "application/json", bytes.NewReader(req)) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("addRelation returned non success status code (%d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func getBanner(slot, group int) (*repository.Banner, error) {
	reqData := struct {
		SlotID  int `json:"slot"`
		GroupID int `json:"group"`
	}{
		SlotID:  slot,
		GroupID: group,
	}

	req, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://127.0.0.1:8088/get_banner", "application/json", bytes.NewReader(req)) //nolint:noctx
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get banners")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	b := repository.Banner{}
	if err := json.Unmarshal(body, &b); err != nil {
		return nil, err
	}

	return &b, nil
}

func TestBanners(t *testing.T) {
	g, err := addGroup("group1")
	assert.NoError(t, err)
	assert.Equal(t, "group1", g.Description)

	b1, err := addBanner("https://mybanner.com/banner1", "banner1")
	assert.NoError(t, err)
	assert.Equal(t, "https://mybanner.com/banner1", b1.URL)
	assert.Equal(t, "banner1", b1.Description)

	b2, err := addBanner("https://mybanner.com/banner2", "banner2")
	assert.NoError(t, err)
	b3, err := addBanner("https://mybanner.com/banner3", "banner3")
	assert.NoError(t, err)

	s, err := addSlot()
	assert.NoError(t, err)

	assert.NoError(t, addRelation(s.ID, b1.ID))
	assert.NoError(t, addRelation(s.ID, b2.ID))
	assert.NoError(t, addRelation(s.ID, b3.ID))

	b, err := getBanner(s.ID, g.ID)
	assert.NoError(t, err)
	assert.True(t, b.ID == b1.ID || b.ID == b2.ID || b.ID == b3.ID)
}

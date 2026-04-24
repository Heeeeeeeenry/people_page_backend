package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"people-page-backend/internal/config"
)

// AMapPOI 高德地图POI
type AMapPOI struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Location string `json:"location"`
	Type     string `json:"type"`
	Typecode string `json:"typecode"`
	Adcode   string `json:"adcode"`
	Adname   string `json:"adname"`
	Tel      string `json:"tel"`
	Pname    string `json:"pname"`
	Cityname string `json:"cityname"`
}

// AMapGeocodeResult 地理编码结果
type AMapGeocodeResult struct {
	FormattedAddress string `json:"formatted_address"`
	Province         string `json:"province"`
	City             string `json:"city"`
	District         string `json:"district"`
	Street           string `json:"street"`
	Number           string `json:"number"`
	Location         string `json:"location"`
	Adcode           string `json:"adcode"`
	Township         string `json:"township"`
}

// AMapRegeocodeResult 逆地理编码结果
type AMapRegeocodeResult struct {
	FormattedAddress string `json:"formatted_address"`
	Province         string `json:"province"`
	City             string `json:"city"`
	District         string `json:"district"`
	Street           string `json:"street"`
	Township         string `json:"township"`
	Adcode           string `json:"adcode"`
}

// SearchPOI POI搜索
func SearchPOI(keywords, city string) ([]AMapPOI, error) {
	cfg := config.AppConfig.Map
	params := url.Values{
		"key":      {cfg.AmapKey},
		"keywords": {keywords},
		"city":     {city},
		"output":   {"JSON"},
	}

	resp, err := http.Get(cfg.AmapPoiURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, nil
	}

	poisRaw, _ := result["pois"].([]interface{})
	var pois []AMapPOI
	for _, p := range poisRaw {
		poi, _ := p.(map[string]interface{})
		pois = append(pois, AMapPOI{
			ID:       getStr(poi, "id"),
			Name:     getStr(poi, "name"),
			Address:  getStr(poi, "address"),
			Location: getStr(poi, "location"),
			Type:     getStr(poi, "type"),
			Typecode: getStr(poi, "typecode"),
			Adcode:   getStr(poi, "adcode"),
			Adname:   getStr(poi, "adname"),
			Tel:      getStr(poi, "tel"),
			Pname:    getStr(poi, "pname"),
			Cityname: getStr(poi, "cityname"),
		})
	}
	return pois, nil
}

// GeocodeAddress 地理编码
func GeocodeAddress(address, city string) (*AMapGeocodeResult, error) {
	cfg := config.AppConfig.Map
	params := url.Values{
		"key":     {cfg.AmapKey},
		"address": {address},
		"city":    {city},
		"output":  {"JSON"},
	}

	resp, err := http.Get(cfg.AmapGeocodeURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, fmt.Errorf("高德API返回状态异常")
	}

	geocodes, _ := result["geocodes"].([]interface{})
	if len(geocodes) == 0 {
		return nil, nil
	}

	loc, _ := geocodes[0].(map[string]interface{})
	return &AMapGeocodeResult{
		FormattedAddress: getStr(loc, "formatted_address"),
		Province:         getStr(loc, "province"),
		City:             getStr(loc, "city"),
		District:         getStr(loc, "district"),
		Street:           getStr(loc, "street"),
		Number:           getStr(loc, "number"),
		Location:         getStr(loc, "location"),
		Adcode:           getStr(loc, "adcode"),
		Township:         getStr(loc, "township"),
	}, nil
}

// RegeocodeLocation 逆地理编码
func RegeocodeLocation(longitude, latitude float64) (*AMapRegeocodeResult, error) {
	cfg := config.AppConfig.Map
	params := url.Values{
		"key":        {cfg.AmapKey},
		"location":   {fmt.Sprintf("%f,%f", longitude, latitude)},
		"extensions": {"all"},
		"output":     {"JSON"},
	}

	resp, err := http.Get(cfg.AmapRegeocodeURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, fmt.Errorf("高德API返回状态异常")
	}

	regeocode, _ := result["regeocode"].(map[string]interface{})
	if regeocode == nil {
		return nil, nil
	}

	comp, _ := regeocode["addressComponent"].(map[string]interface{})
	return &AMapRegeocodeResult{
		FormattedAddress: getStr(regeocode, "formatted_address"),
		Province:         getStr(comp, "province"),
		City:             getStr(comp, "city"),
		District:         getStr(comp, "district"),
		Street:           getStr(comp, "street"),
		Township:         getStr(comp, "township"),
		Adcode:           getStr(comp, "adcode"),
	}, nil
}

// SearchPOIAround 周边搜索
func SearchPOIAround(location, keywords string, radius int) ([]AMapPOI, error) {
	cfg := config.AppConfig.Map
	params := url.Values{
		"key":      {cfg.AmapKey},
		"location": {location},
		"keywords": {keywords},
		"radius":   {fmt.Sprintf("%d", radius)},
		"output":   {"JSON"},
	}

	resp, err := http.Get(cfg.AmapAroundURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, nil
	}

	poisRaw, _ := result["pois"].([]interface{})
	var pois []AMapPOI
	for _, p := range poisRaw {
		poi, _ := p.(map[string]interface{})
		pois = append(pois, AMapPOI{
			ID:       getStr(poi, "id"),
			Name:     getStr(poi, "name"),
			Address:  getStr(poi, "address"),
			Location: getStr(poi, "location"),
			Type:     getStr(poi, "type"),
			Adcode:   getStr(poi, "adcode"),
			Adname:   getStr(poi, "adname"),
		})
	}
	return pois, nil
}

// GetInputTips 输入提示
func GetInputTips(keywords, city string) ([]map[string]interface{}, error) {
	cfg := config.AppConfig.Map
	params := url.Values{
		"key":      {cfg.AmapKey},
		"keywords": {keywords},
		"city":     {city},
		"output":   {"JSON"},
	}

	resp, err := http.Get(cfg.AmapInputTipsURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result["status"] != "1" {
		return nil, nil
	}

	tipsRaw, _ := result["tips"].([]interface{})
	var tips []map[string]interface{}
	for _, t := range tipsRaw {
		tip, _ := t.(map[string]interface{})
		tips = append(tips, map[string]interface{}{
			"id":       getStr(tip, "id"),
			"name":     getStr(tip, "name"),
			"address":  getStr(tip, "address"),
			"location": getStr(tip, "location"),
			"adcode":   getStr(tip, "adcode"),
			"district": getStr(tip, "district"),
		})
	}
	return tips, nil
}

func getStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

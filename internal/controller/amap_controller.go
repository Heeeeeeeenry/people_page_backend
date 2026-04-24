package controller

import (
	"net/http"
	"strconv"
	"people-page-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// SearchPOI POI搜索
func SearchPOI(c *gin.Context) {
	keywords := c.Query("keywords")
	city := c.Query("city")

	if keywords == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keywords参数必填"})
		return
	}

	pois, err := service.SearchPOI(keywords, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "1",
		"pois":   pois,
	})
}

// Geocode 地理编码
func Geocode(c *gin.Context) {
	address := c.Query("address")
	city := c.Query("city")

	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address参数必填"})
		return
	}

	result, err := service.GeocodeAddress(address, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "1",
		"geocodes": []interface{}{result},
	})
}

// Regeocode 逆地理编码
func Regeocode(c *gin.Context) {
	lonStr := c.Query("longitude")
	latStr := c.Query("latitude")

	lon, err1 := strconv.ParseFloat(lonStr, 64)
	lat, err2 := strconv.ParseFloat(latStr, 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "longitude/latitude格式错误"})
		return
	}

	result, err := service.RegeocodeLocation(lon, lat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "1",
		"regeocode": gin.H{
			"addressComponent": result,
			"formatted_address": result.FormattedAddress,
		},
	})
}

// SearchPOIAround 周边搜索
func SearchPOIAround(c *gin.Context) {
	location := c.Query("location")
	keywords := c.Query("keywords")
	radiusStr := c.Query("radius")

	radius := 1000
	if r, err := strconv.Atoi(radiusStr); err == nil && r > 0 {
		radius = r
	}

	pois, err := service.SearchPOIAround(location, keywords, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "1",
		"pois":   pois,
	})
}

// GetInputTips 输入提示
func GetInputTips(c *gin.Context) {
	keywords := c.Query("keywords")
	city := c.Query("city")

	if keywords == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keywords参数必填"})
		return
	}

	tips, err := service.GetInputTips(keywords, city)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "1",
		"tips":   tips,
	})
}

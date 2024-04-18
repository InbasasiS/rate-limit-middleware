package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rate-limit/app/cache"
)

type MetricsController interface {
	GetCurrentHits(http.ResponseWriter, *http.Request)
}

type metricsController struct {
	cache cache.Cache
}

const (
	TimeWindowKey = "TimeWindow::"
	NoOfHitsKey   = "NoOfHits::"
)

type ApiHits struct {
	IpAddress  string
	TimeWindow string
	ApiPath    string
	NoOfHits   string
}

func NewMetricsController() MetricsController {
	return metricsController{
		cache: cache.NewCacheClient(),
	}
}

func (rlc metricsController) GetCurrentHits(w http.ResponseWriter, r *http.Request) {
	ipAddress := r.RemoteAddr
	requestedRoute := r.URL.Path
	queryParams := r.URL.Query()
	appId := queryParams.Get("appId")

	timeWindowCacheKey := getCacheKeyForTimeWindow(requestedRoute, appId, ipAddress)
	startTime, _ := rlc.cache.Get(timeWindowCacheKey)
	timeEpoch, _ := strconv.ParseInt(startTime, 10, 64)
	hitsCacheKey := getCacheKeyForHits(timeWindowCacheKey, startTime)
	hits, _ := rlc.cache.Get(hitsCacheKey)
	apiHits := ApiHits{
		IpAddress:  ipAddress,
		TimeWindow: time.Unix(timeEpoch, 0).Format("2006-01-02T15:04:05Z"),
		NoOfHits:   hits,
		ApiPath:    requestedRoute,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiHits)
}

func getCacheKeyForHits(timeWindowCacheKey string, startTime string) string {
	return timeWindowCacheKey + "::" + startTime
}

func getCacheKeyForTimeWindow(requestedRoute string, appId string, ipAddress string) string {
	if appId != "" {
		return TimeWindowKey + requestedRoute + "::" + appId
	}
	return TimeWindowKey + requestedRoute + "::" + ipAddress
}

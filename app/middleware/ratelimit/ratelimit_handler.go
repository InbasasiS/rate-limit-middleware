package ratelimit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rate-limit/app/cache"
	"github.com/rate-limit/config"
)

const (
	TimeWindowKey = "TimeWindow::"
	NoOfHitsKey   = "NoOfHits::"
)

func RatelimitHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if shouldServeRequest(w, r) {
			next.ServeHTTP(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode("409 Too Many Requests")
		}
	}
	return http.HandlerFunc(fn)
}

func shouldServeRequest(w http.ResponseWriter, r *http.Request) bool {
	ipAddress := r.RemoteAddr
	requestedRoute := r.URL.Path
	queryParams := r.URL.Query()
	appId := queryParams.Get("appId")
	cache := cache.NewCacheClient()
	rateLimitConfig := getRateLimitConfig(appId, ipAddress)
	if rateLimitConfig == nil {
		fmt.Println("Rate limit not configured")
		return true
	}

	timeWindowCacheKey := getCacheKeyForTimeWindow(requestedRoute, appId, ipAddress)
	startTime, err := cache.Get(timeWindowCacheKey)
	if err != nil {
		fmt.Println("Error fetching time window, ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if startTime == "" || timeElasped(startTime, rateLimitConfig) {
		fmt.Println("Time elasped, Resetting the time window")
		startTime = fmt.Sprintf("%d", time.Now().Unix())
		tPlusIntervalDuration := time.Duration(rateLimitConfig.Interval.Value) * getTimeUnitMultiplier(string(rateLimitConfig.Interval.TimeUnit))
		err := cache.Set(timeWindowCacheKey, startTime, tPlusIntervalDuration)
		if err != nil {
			fmt.Println("Error creating time window, ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	hitsCacheKey := getCacheKeyForHits(timeWindowCacheKey, startTime)
	hitCount, err := cache.Incr(hitsCacheKey)
	if err != nil {
		fmt.Println("Error fetching hit count, ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	if hitCount == 1 {
		// Set Expiry for the key initially
		hitsCacheKey := getCacheKeyForHits(timeWindowCacheKey, startTime)
		tPlusIntervalDuration := time.Duration(rateLimitConfig.Interval.Value) * getTimeUnitMultiplier(string(rateLimitConfig.Interval.TimeUnit))
		cache.Expire(hitsCacheKey, tPlusIntervalDuration)
	}

	fmt.Println("timeWindowCacheKey ", timeWindowCacheKey)
	fmt.Println("hitsCacheKey, ", hitsCacheKey)
	if hitCount > int64(rateLimitConfig.Threshold) {
		fmt.Printf("Total Hits %d, greater than threshold %d", hitCount, rateLimitConfig.Threshold)
		return false
	}

	fmt.Printf("Serving request, current hits/threshold %d/%d", hitCount, rateLimitConfig.Threshold)
	return true
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

func timeElasped(timeWindow string, rateLimit *config.RateLimitConfig) bool {
	timeEpoch, _ := strconv.ParseInt(timeWindow, 10, 64)
	t := time.Unix(timeEpoch, 0)
	tPlusInterval := t.Add(time.Duration(rateLimit.Interval.Value) * getTimeUnitMultiplier(string(rateLimit.Interval.TimeUnit)))
	return time.Now().After(tPlusInterval)
}

func getTimeUnitMultiplier(timeUnit string) time.Duration {
	switch timeUnit {
	case "MINUTE":
		return time.Minute
	case "HOUR":
		return time.Hour
	case "DAY":
		return time.Hour * 24
	}
	return time.Second
}

func getRateLimitConfig(appId string, addr string) *config.RateLimitConfig {
	rateLimitConfig := config.GetRateLimitConfig()
	rateLimitByAppId := rateLimitConfig[appId]
	rateLimitByAddr := rateLimitConfig[addr]
	if rateLimitByAppId != nil {
		return rateLimitByAppId
	}
	if rateLimitByAddr != nil {
		return rateLimitByAddr
	}
	return rateLimitConfig["default"]
}

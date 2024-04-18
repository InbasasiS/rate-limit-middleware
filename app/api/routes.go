package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rate-limit/app/api/controller"
	"github.com/rate-limit/app/middleware/ratelimit"
)

func GetRoutes() *mux.Router {
	router := mux.NewRouter()
	r1 := router.PathPrefix("/api/metrics").Subrouter()

	metricsController := controller.NewMetricsController()
	r1.HandleFunc("/hits", metricsController.GetCurrentHits).Methods(http.MethodGet)
	r1.Use(ratelimit.RatelimitHandler)
	return router
}

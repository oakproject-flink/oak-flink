package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/oakproject-flink/oak-flink/oak-server/web/templates/pages"
)

// Dashboard renders the main dashboard page
func Dashboard(c echo.Context) error {
	// TODO: Fetch real data from services
	stats := map[string]interface{}{
		"totalJobs":    24,
		"runningJobs":  18,
		"failingJobs":  2,
		"totalClusters": 3,
	}

	return pages.Dashboard(stats).Render(c.Request().Context(), c.Response())
}

// Jobs renders the jobs list page
func Jobs(c echo.Context) error {
	return pages.Jobs().Render(c.Request().Context(), c.Response())
}

// Clusters renders the clusters page
func Clusters(c echo.Context) error {
	return pages.Clusters().Render(c.Request().Context(), c.Response())
}

// Metrics renders the metrics page
func Metrics(c echo.Context) error {
	return pages.Metrics().Render(c.Request().Context(), c.Response())
}
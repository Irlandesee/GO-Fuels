package routers

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo) {
	fuelsGroup := e.Group("/fuels")
	locationGroup := e.Group("/location")
	userGroup := e.Group("/user")
	brandGroup := e.Group("/brand")
	RegisterFuelRoutes(e, fuelsGroup)
	RegisterLocationRoutes(e, locationGroup)
	RegisterUserRoutes(e, userGroup)
	RegisterBrandRoutes(e, brandGroup)
}

func RegisterFuelRoutes(e *echo.Echo, g *echo.Group, h *Handler) {
	// Fuel Types (MongoDB)
	g.POST("/types", h.CreateFuelType)
	g.GET("/types", h.GetAllFuelTypes)
	g.GET("/types/:fuel_key", h.GetFuelTypeByKey)
	g.GET("/types/category/:category", h.GetFuelTypesByCategory)

	// Fuel Data/Prices (PostgreSQL)
	g.POST("/data", h.CreateFuelData)
	g.PUT("/data", h.UpsertFuelData)
	g.GET("/data/location/:location_key", h.GetFuelDataByLocation)
	g.GET("/data/location/:location_key/fuel/:fuel_key", h.GetFuelDataByLocationAndFuel)
	g.GET("/data/history", h.GetPriceHistory)
	g.GET("/data/average/:fuel_key", h.GetAveragePrice)
	g.GET("/data/latest/:fuel_key", h.GetLatestPrices)
	g.PATCH("/data/price", h.UpdateFuelPrice)
}

func RegisterLocationRoutes(e *echo.Echo, g *echo.Group, h *Handler) {
	g.POST("", h.CreateLocation)
	g.GET("/:location_key", h.GetLocationByKey)
	g.PUT("/:location_key", h.UpdateLocation)
	g.GET("/brand/:brand_key", h.GetLocationsByBrand)
	g.GET("/nearby", h.GetNearbyLocations)

	// Combined queries
	g.GET("/:location_key/prices", h.GetLocationWithPrices)
	g.GET("/nearby/prices", h.GetNearbyLocationsWithPrices)
}

func RegisterUserRoutes(e *echo.Echo, g *echo.Group, h *Handler) {
	g.POST("/preferences", h.CreateUserPreference)
	g.GET("/preferences/:user_id", h.GetUserPreference)
	g.PUT("/preferences/:user_id", h.UpdateUserPreference)
}

func RegisterBrandRoutes(e *echo.Echo, g *echo.Group, h *Handler) {
	g.POST("", h.CreateBrand)
	g.GET("/:brand_key", h.GetBrandByKey)
	g.GET("", h.GetAllBrands)
	g.PUT("/:brand_key", h.UpdateBrand)
}

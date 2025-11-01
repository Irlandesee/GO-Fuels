package routers

import (
	"Irlandesee/GO-Fuels/src/utils/dbHandlers"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, dbHandler *dbHandlers.Handler) {
	// Create handlers
	fuelHandler := NewFuelHandler(dbHandler)
	locationHandler := NewLocationHandler(dbHandler)
	userHandler := NewUserHandler(dbHandler)
	brandHandler := NewBrandHandler(dbHandler)

	// Create route groups
	fuelsGroup := e.Group("/fuels")
	locationGroup := e.Group("/location")
	userGroup := e.Group("/user")
	brandGroup := e.Group("/brand")

	// Register routes
	RegisterFuelRoutes(fuelsGroup, fuelHandler)
	RegisterLocationRoutes(locationGroup, locationHandler)
	RegisterUserRoutes(userGroup, userHandler)
	RegisterBrandRoutes(brandGroup, brandHandler)
}

func RegisterFuelRoutes(g *echo.Group, h *FuelHandler) {
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

func RegisterLocationRoutes(g *echo.Group, h *LocationHandler) {
	g.POST("", h.CreateLocation)
	g.GET("/:location_key", h.GetLocationByKey)
	g.PUT("/:location_key", h.UpdateLocation)
	g.GET("/brand/:brand_key", h.GetLocationsByBrand)
	g.GET("/nearby", h.GetNearbyLocations)

	// Combined queries
	g.GET("/:location_key/prices", h.GetLocationWithPrices)
	g.GET("/nearby/prices", h.GetNearbyLocationsWithPrices)
}

func RegisterUserRoutes(g *echo.Group, h *UserHandler) {
	g.POST("/preferences", h.CreateUserPreference)
	g.GET("/preferences/:user_id", h.GetUserPreference)
	g.PUT("/preferences/:user_id", h.UpdateUserPreference)
}

func RegisterBrandRoutes(g *echo.Group, h *BrandHandler) {
	g.POST("", h.CreateBrand)
	g.GET("/:brand_key", h.GetBrandByKey)
	g.GET("", h.GetAllBrands)
	g.PUT("/:brand_key", h.UpdateBrand)
}

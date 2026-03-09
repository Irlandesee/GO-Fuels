package utils

import (
	"fmt"
	"strings"
)

// GenerateFuelKey creates a consistent fuel key from category and type.
// Example: "gasoline_95", "diesel_premium", "electric_fast"
func GenerateFuelKey(category, fuelType string) string {
	category = strings.ToLower(strings.TrimSpace(category))
	fuelType = strings.ToLower(strings.TrimSpace(fuelType))

	// Replace spaces and special chars with underscores
	replacer := strings.NewReplacer(" ", "_", "-", "_", "/", "_")
	category = replacer.Replace(category)
	fuelType = replacer.Replace(fuelType)

	if fuelType == "" {
		return category
	}
	return fmt.Sprintf("%s_%s", category, fuelType)
}

// GenerateLocationKey creates a consistent location key.
// Example: "shell_main_street_123", "eni_highway_a1_km45"
func GenerateLocationKey(brand, address string) string {
	brand = strings.ToLower(strings.TrimSpace(brand))
	address = strings.ToLower(strings.TrimSpace(address))

	replacer := strings.NewReplacer(" ", "_", "-", "_", "/", "_", ",", "", ".", "")
	brand = replacer.Replace(brand)
	address = replacer.Replace(address)

	return fmt.Sprintf("%s_%s", brand, address)
}

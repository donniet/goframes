package frames

const (
	Gravity = 32.174048554 // ft/sec^2
)

func WindPressure(airDensity, windSpeed float64) float64 {
	return 0.5 * airDensity / Gravity * windSpeed * windSpeed
}

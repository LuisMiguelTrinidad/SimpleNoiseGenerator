package terrain

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// InterpolateHeight finds the terrain height at any continuous point using bilinear interpolation
func InterpolateHeight(heightmap [][]float64, x, y float64) float64 {
	// Convert floating-point coordinates to integer and fractional parts
	ix := int(x)
	iy := int(y)
	fx := x - float64(ix) // Fractional part of x
	fy := y - float64(iy) // Fractional part of y

	// Get height and width of the heightmap
	height := len(heightmap)
	width := len(heightmap[0])

	// Ensure coordinates stay within bounds
	ix0 := max(0, min(ix, width-1))
	ix1 := max(0, min(ix+1, width-1))
	iy0 := max(0, min(iy, height-1))
	iy1 := max(0, min(iy+1, height-1))

	// Get heights at the four surrounding grid points
	h00 := heightmap[iy0][ix0] // Top-left
	h10 := heightmap[iy0][ix1] // Top-right
	h01 := heightmap[iy1][ix0] // Bottom-left
	h11 := heightmap[iy1][ix1] // Bottom-right

	// Perform bilinear interpolation
	h0 := h00*(1-fx) + h10*fx // Interpolate along top edge
	h1 := h01*(1-fx) + h11*fx // Interpolate along bottom edge
	return h0*(1-fy) + h1*fy  // Interpolate between top and bottom
}

// ComputeGradient calculates the slope direction at a specific point
func ComputeGradient(heightmap [][]float64, x, y float64, epsilon float64) (float64, float64) {
	// Use central difference method to estimate partial derivatives
	gx := (InterpolateHeight(heightmap, x+epsilon, y) - InterpolateHeight(heightmap, x-epsilon, y)) / (2 * epsilon)
	gy := (InterpolateHeight(heightmap, x, y+epsilon) - InterpolateHeight(heightmap, x, y-epsilon)) / (2 * epsilon)
	return gx, gy
}

// Parameters for erosion simulation
type ErosionParams struct {
	MaxSteps         int     // Maximum lifetime of each droplet
	Inertia          float64 // How much a droplet maintains its direction
	SedimentCapacity float64 // How much sediment a droplet can carry
	ErosionRate      float64 // How quickly droplets pick up sediment
	DepositionRate   float64 // How quickly droplets deposit sediment
	EvaporationRate  float64 // How quickly water evaporates
	Gravity          float64 // Affects droplet velocity
	MinSlope         float64 // Minimum slope for movement
	CellSize         float64 // Scale factor for movement distance
}

// ApplyErosion simulates hydraulic erosion by running multiple water droplets across the terrain
func ApplyErosion(heightmap [][]float64, numDroplets int, params ErosionParams) [][]float64 {
	startTotal := time.Now()
	fmt.Printf("Iniciando simulación de erosión hidráulica (%d gotas)...\n", numDroplets)

	// Create a copy of the heightmap to avoid modifying the original
	startCopy := time.Now()
	height := len(heightmap)
	width := len(heightmap[0])
	result := make([][]float64, height)
	for i := range result {
		result[i] = make([]float64, width)
		copy(result[i], heightmap[i])
	}
	fmt.Printf("  ├─ Copia del mapa: %.3f ms\n",
		float64(time.Since(startCopy).Microseconds())/1000)

	// Estadísticas
	totalSteps := 0
	maxSteps := 0
	dropletsOffMap := 0
	dropletsEvaporated := 0
	dropletsNoDirection := 0
	totalDeposited := 0.0
	totalEroded := 0.0

	// Simulate each water droplet
	startDroplets := time.Now()
	reportInterval := numDroplets / 10 // Reportar progreso cada 10%
	if reportInterval < 1 {
		reportInterval = 1
	}

	for d := 0; d < numDroplets; d++ {
		if d > 0 && d%reportInterval == 0 {
			pctComplete := float64(d) / float64(numDroplets) * 100
			timeElapsed := time.Since(startDroplets)
			timeEstimated := time.Duration(float64(timeElapsed) / (float64(d) / float64(numDroplets)))
			timeRemaining := timeEstimated - timeElapsed

			fmt.Printf("  │  ├─ %.1f%% completado - Tiempo restante: %.1f s (%.0f gotas/s)\n",
				pctComplete, timeRemaining.Seconds(),
				float64(d)/timeElapsed.Seconds())
		}

		// Random starting position for the droplet
		x := rand.Float64() * float64(width-1)
		y := rand.Float64() * float64(height-1)

		// Initial movement direction, velocity, water volume, and sediment
		dirX, dirY := 0.0, 0.0
		velocity := 0.0
		water := 1.0
		sediment := 0.0
		steps := 0
		dropletDeposited := 0.0
		dropletEroded := 0.0

		// Simulate each step of the droplet's lifetime
		for step := 0; step < params.MaxSteps; step++ {
			steps++

			// Calculate gradient at current position
			gx, gy := ComputeGradient(result, x, y, 1e-5)
			slope := math.Sqrt(gx*gx + gy*gy)

			// If slope is too shallow, water wouldn't flow
			if slope < params.MinSlope {
				gx, gy, slope = 0.0, 0.0, 0.0
			}

			// Calculate movement direction with inertia
			dirX = dirX*params.Inertia + gx*(1-params.Inertia)
			dirY = dirY*params.Inertia + gy*(1-params.Inertia)
			dirLength := math.Hypot(dirX, dirY)

			// If no direction, droplet stops moving
			if dirLength == 0 {
				dropletsNoDirection++
				break
			}

			// Normalize direction vector
			dirX /= dirLength
			dirY /= dirLength

			// Calculate new position
			newX := x + dirX*params.CellSize
			newY := y + dirY*params.CellSize

			// Stop if droplet flows off the map
			if newX < 0 || newX >= float64(width) || newY < 0 || newY >= float64(height) {
				dropletsOffMap++
				break
			}

			// Calculate height difference between old and new position
			oldHeight := InterpolateHeight(result, x, y)
			newHeight := InterpolateHeight(result, newX, newY)
			deltaH := newHeight - oldHeight

			// Calculate sediment capacity based on slope and velocity
			capacity := math.Max(-deltaH, 0.0) * velocity * params.SedimentCapacity
			capacity = math.Max(capacity, params.MinSlope)

			// Handle deposition (when carrying too much sediment or going uphill)
			if sediment > capacity || deltaH > 0 {
				depositAmount := math.Min((sediment-capacity)*params.DepositionRate, sediment)
				sediment -= depositAmount
				dropletDeposited += depositAmount
				totalDeposited += depositAmount

				ix, iy := int(x), int(y)
				fx, fy := x-float64(ix), y-float64(iy)

				// Distribute deposited sediment to surrounding cells
				for di := 0; di <= 1; di++ {
					for dj := 0; dj <= 1; dj++ {
						// Calculate bilinear weight
						wi := 0.0
						if di == 0 {
							wi = (1 - fx)
						} else {
							wi = fx
						}

						if dj == 0 {
							wi *= (1 - fy)
						} else {
							wi *= fy
						}

						i, j := ix+di, iy+dj
						if i >= 0 && i < width && j >= 0 && j < height {
							result[j][i] += depositAmount * wi
						}
					}
				}
			} else {
				// Handle erosion (when carrying less than capacity and going downhill)
				erosionAmount := math.Min((capacity-sediment)*params.ErosionRate, -deltaH)
				erosionAmount = math.Max(erosionAmount, 0)

				ix, iy := int(x), int(y)
				fx, fy := x-float64(ix), y-float64(iy)
				totalWeight := 0.0

				// Erode from surrounding cells
				for di := 0; di <= 1; di++ {
					for dj := 0; dj <= 1; dj++ {
						// Calculate bilinear weight
						wi := 0.0
						if di == 0 {
							wi = (1 - fx)
						} else {
							wi = fx
						}

						if dj == 0 {
							wi *= (1 - fy)
						} else {
							wi *= fy
						}

						i, j := ix+di, iy+dj
						if i >= 0 && i < width && j >= 0 && j < height {
							// Limit erosion to prevent negative heights
							erode := math.Min(erosionAmount*wi, result[j][i])
							result[j][i] -= erode
							sediment += erode
							totalWeight += wi
							dropletEroded += erode
							totalEroded += erode
						}
					}
				}

				// Account for potential cells outside the map
				if totalWeight > 0 {
					additionalSediment := erosionAmount * (1 - totalWeight)
					sediment += additionalSediment
					totalEroded += additionalSediment
				}
			}

			// Update droplet properties
			velocity = math.Sqrt(velocity*velocity + deltaH*params.Gravity)
			velocity = math.Max(velocity, 0)
			water *= (1 - params.EvaporationRate)

			// When too much water evaporates, the droplet's journey ends
			if water < 0.01 {
				dropletsEvaporated++
				break
			}

			// Move to new position
			x, y = newX, newY
		}

		totalSteps += steps
		if steps > maxSteps {
			maxSteps = steps
		}
	}

	simulationTime := time.Since(startDroplets)
	fmt.Printf("  ├─ Simulación de erosión: %.3f s (%.0f gotas/s)\n",
		simulationTime.Seconds(), float64(numDroplets)/simulationTime.Seconds())

	// Mostrar estadísticas
	fmt.Printf("  ├─ Estadísticas:\n")
	fmt.Printf("  │  ├─ Pasos promedio por gota: %.1f (máx: %d)\n", float64(totalSteps)/float64(numDroplets), maxSteps)
	fmt.Printf("  │  ├─ Gotas evaporadas: %d (%.1f%%)\n", dropletsEvaporated, float64(dropletsEvaporated)/float64(numDroplets)*100)
	fmt.Printf("  │  ├─ Gotas fuera del mapa: %d (%.1f%%)\n", dropletsOffMap, float64(dropletsOffMap)/float64(numDroplets)*100)
	fmt.Printf("  │  ├─ Gotas sin dirección: %d (%.1f%%)\n", dropletsNoDirection, float64(dropletsNoDirection)/float64(numDroplets)*100)
	fmt.Printf("  │  ├─ Material erosionado: %.1f unidades\n", totalEroded)
	fmt.Printf("  │  └─ Material depositado: %.1f unidades\n", totalDeposited)

	fmt.Printf("  └─ Tiempo total de erosión: %.3f s\n", time.Since(startTotal).Seconds())

	return result
}

// ApplyErosionAndClamp aplica erosión hidráulica y luego asegura que todos los valores
// permanezcan dentro del rango [-1, 1]
func ApplyErosionAndClamp(heightmap [][]float64, numDroplets int, params ErosionParams) [][]float64 {
	startTotal := time.Now()
	fmt.Printf("Iniciando erosión con límites...\n")

	// Aplicar el algoritmo de erosión existente
	startErosion := time.Now()
	result := ApplyErosion(heightmap, numDroplets, params)
	fmt.Printf("  ├─ Tiempo de erosión base: %.3f s\n", time.Since(startErosion).Seconds())

	// Limitar los valores al rango [-1, 1]
	startClamp := time.Now()
	clampedAbove := 0
	clampedBelow := 0

	for i := range result {
		for j := range result[i] {
			if result[i][j] > 1.0 {
				result[i][j] = 1.0
				clampedAbove++
			} else if result[i][j] < -1.0 {
				result[i][j] = -1.0
				clampedBelow++
			}
		}
	}

	totalCells := len(result) * len(result[0])
	fmt.Printf("  ├─ Limitación de valores: %.3f ms\n", float64(time.Since(startClamp).Microseconds())/1000)
	fmt.Printf("  ├─ Celdas limitadas: %d de %d (%.2f%%)\n",
		clampedAbove+clampedBelow, totalCells,
		float64(clampedAbove+clampedBelow)/float64(totalCells)*100)
	fmt.Printf("  │  ├─ Limitadas por arriba (>1.0): %d\n", clampedAbove)
	fmt.Printf("  │  └─ Limitadas por abajo (<-1.0): %d\n", clampedBelow)

	fmt.Printf("  └─ Tiempo total de erosión con límites: %.3f s\n", time.Since(startTotal).Seconds())

	return result
}

// Helper functions for min and max (for Go versions before 1.21)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

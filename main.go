package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"main/terrain"
)

func createNoiseMap(size int, seed int64, octaves int, scale float64) *terrain.Terrain {
	noise := terrain.NewOpenSimplex(seed) // Corrected to opensimplex.New
	terrainObj := terrain.NewTerrain(size)

	// Precompute amplitudeSum, freqs and amps
	freqs := make([]float64, octaves)
	amps := make([]float64, octaves)
	persistence := 0.5
	amplitudeSum := 0.0
	for i := range octaves {
		freqs[i] = math.Pow(2, float64(i))
		amps[i] = math.Pow(persistence, float64(i))
		amplitudeSum += amps[i]
	}

	// Generate initial heightmap - Sequential version
	for y := range size {
		for x := range size {
			var total float64
			for i := range octaves {
				nx := float64(x) / scale * freqs[i]
				ny := float64(y) / scale * freqs[i]
				total += noise.Eval2(nx, ny) * amps[i]
			}
			value := total / amplitudeSum
			terrainObj.Heightmap[x][y] = (1 + ((0.1*value+0.9*(math.Pow(value, 3)))/2)*float64(size))
		}
	}
	return terrainObj
}

func applyErosion(t *terrain.Terrain, iterations int, dropletCount int,
	erosionRadius float64, inertia float64, sedimentCapacity float64,
	evaporation float64, gravity float64) {

	t.ApplyErosion(terrain.ErosionParams{
		Iterations:       iterations,
		DropletCount:     dropletCount,
		ErosionRadius:    erosionRadius,
		Inertia:          inertia,
		SedimentCapacity: sedimentCapacity,
		Evaporation:      evaporation,
		Gravity:          gravity,
	})
}

func main() {
	start_time := time.Now()

	const (
		size             = 256
		scale            = 256
		octaves          = 12
		seed             = 0
		iterations       = 10
		dropletCount     = 2000000
		erosionRadius    = 4
		inertia          = 0.1
		sedimentCapacity = 0.01
		evaporation      = 0.1
		gravity          = 4
	)

	// Create output directory if not exists
	outputDir := "images"
	os.MkdirAll(outputDir, 0755)

	fmt.Printf("%v Generating terrain...\n", time.Since(start_time))

	// Create noise map
	terrainObj := createNoiseMap(size, seed, octaves, scale)

	fmt.Printf("%v Applying erosion...\n", time.Since(start_time))

	// Apply hydraulic erosion
	applyErosion(terrainObj, iterations, dropletCount, erosionRadius,
		inertia, sedimentCapacity, evaporation, gravity)

	fmt.Printf("%v Post-processing terrain...\n", time.Since(start_time))

	fmt.Printf("%v Saving terrain...\n", time.Since(start_time))

	// Create outputs
	vertices, vertexColors := terrainObj.GenerateVertices()
	faces := terrainObj.GenerateFaces()
	terrain.SavePLY(filepath.Join(outputDir, "eroded_terrain.ply"), vertices, faces, vertexColors)

	fmt.Printf("%v Done!\n", time.Since(start_time))
}

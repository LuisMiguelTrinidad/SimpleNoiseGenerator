package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"main/terrain"
)

func createNoiseMap(mapseed int64, mapSize int, mapScale float64, mapOctaves int, smoothingFunction func(float64) float64) *terrain.Terrain {
	noise := terrain.NewOpenSimplex(mapseed)
	terrainObj := terrain.NewTerrain(mapSize)

	// Precompute amplitudeSum, freqs and amps
	freqs := make([]float64, mapOctaves)
	amps := make([]float64, mapOctaves)
	persistence := 0.5
	amplitudeSum := 0.0
	for i := range mapOctaves {
		freqs[i] = math.Pow(2, float64(i))
		amps[i] = math.Pow(persistence, float64(i))
		amplitudeSum += amps[i]
	}

	// Generate initial heightmap - Sequential version
	for y := range mapSize {
		for x := range mapSize {
			var total float64
			for i := range mapOctaves {
				nx := float64(x) / mapScale * freqs[i]
				ny := float64(y) / mapScale * freqs[i]
				total += noise.Eval2(nx, ny) * amps[i]
			}
			terrainObj.Heightmap[x][y] = total / amplitudeSum
		}
	}

	for y := range mapSize {
		for x := range mapSize {
			terrainObj.Heightmap[x][y] = smoothingFunction(terrainObj.Heightmap[x][y]) * 256
		}
	}
	return terrainObj
}

// Aplica erosión hidráulica usando la biblioteca rainfall
func applyErosion(t *terrain.Terrain, iterations int, dropletCount int, inertia float64,
	sedimentCapacity float64, evaporation float64, gravity float64, seed int64) error {
	params := terrain.ErosionParams{
		Iterations:       iterations,
		DropletCount:     dropletCount,
		Inertia:          inertia,
		SedimentCapacity: sedimentCapacity,
		Evaporation:      evaporation,
		Gravity:          gravity,
		Seed:             seed,
	}
	t.ApplyErosion(params)
	return nil
}

func main() {
	start_time := time.Now()

	const (
		MapSize    = 512
		MapScale   = 2048
		MapOctaves = 12
		MapSeed    = 0
	)

	MapSmoothingFunction := func(height float64) float64 {
		for range 1 {
			height = 0.75*math.Copysign((math.Sin(math.Pi*height-math.Pi/2)/2+0.5), height) + 0.25*height
		}
		return height
	}

	const (
		ErosionIterations       = 100
		ErosionDropletCount     = 200000
		ErosionInertia          = 0.05
		ErosionSedimentCapacity = 0.3
		ErosionEvaporation      = 1 / MapSize
		ErosionGravity          = 9.8
		ErosionSeed             = 0
	)

	// Create output directory if not exists
	const (
		imageDir = "images"
		meshDir  = "meshes"
	)

	os.MkdirAll(imageDir, 0755)
	os.MkdirAll(meshDir, 0755)

	fmt.Printf("%v Generating terrain...\n", time.Since(start_time))

	terrainObj := createNoiseMap(MapSeed, MapSize, MapScale, MapOctaves, MapSmoothingFunction)

	fmt.Printf("%v Aplicando erosión con Rainfall...\n", time.Since(start_time))

	const num = 10
	for i := range num {
		// Apply erosion and save result
		applyErosion(terrainObj, ErosionIterations, ErosionDropletCount, ErosionInertia,
			ErosionSedimentCapacity, ErosionEvaporation, ErosionGravity, ErosionSeed)

		// Generate and save mesh
		vertices, vertexColors := terrainObj.GenerateVertices()
		faces := terrainObj.GenerateFaces()
		plyPath := filepath.Join(meshDir, fmt.Sprintf("eroded_terrain_%d.ply", i))
		terrain.SavePLY(plyPath, vertices, faces, vertexColors)

		// Render image with Python script
		imgPath := filepath.Join(imageDir, fmt.Sprintf("terrain_render_%d.png", i))
		cmd := exec.Command("python", "terrain/run.py",
			"-i", plyPath,
			"-o", imgPath,
			"-c", "terrain",
			"-b", "black",
			"-v", "isometric")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running Python script: %v\n", err)
		} else {
			fmt.Printf("%v, %v/%vSaved image to %s\n", time.Since(start_time), i, num, imgPath)
		}
	}

}

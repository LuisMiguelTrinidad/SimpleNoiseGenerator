package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"main/terrain"
)

func main() {
	const (
		MapSize    = 1024
		MapScale   = 2048
		MapHeight  = 256
		MapOctaves = 12
		MapSeed    = 3421
	)

	// Create output directory if not exists
	const (
		imageDir = "images"
		meshDir  = "meshes"
	)

	ErosionParams := terrain.ErosionParams{
		MaxSteps:         100,
		Inertia:          0.05,
		SedimentCapacity: 3.0,
		ErosionRate:      0.3,
		DepositionRate:   0.3,
		EvaporationRate:  0.01,
		Gravity:          9.8,
		MinSlope:         0.01,
		CellSize:         1.0,
	}
	ErosionDropletCount := 200000

	os.MkdirAll(imageDir, 0755)
	os.MkdirAll(meshDir, 0755)
	start_time := time.Now()
	fmt.Printf("\nInicio de generación: %s\n", time.Now().Format("15:04:05"))

	Smoother := func(height float64) float64 {
		height = terrain.GreatPlains(height)
		height = terrain.Plateau(height, 0.75)
		height = terrain.Plateau(height, 0.75)
		height = terrain.Molone(height, 0.8)
		height = terrain.GreatPlains(height)
		height = terrain.Plateau(height, 0.2)
		return height
	}

	noiseStart := time.Now()
	heightmap := terrain.CreateNoiseMap(MapSeed, MapSize, MapScale, MapOctaves, Smoother)
	fmt.Printf("\nGeneración de mapa de ruido: %.3f segundos\n", time.Since(noiseStart).Seconds())

	const num = 10
	for i := range num {
		iterStart := time.Now()
		fmt.Printf("\n--- Iteración %d/%d ---\n", i+1, num)

		// Aplicar erosión con el nuevo enfoque (solo necesita el número de gotas)
		scaleStart := time.Now()
		// Create a scaled copy of heightmap from [-1,1] to [0,256]
		scaledHeightmap := make([][]float64, len(heightmap))
		for y := range heightmap {
			scaledHeightmap[y] = make([]float64, len(heightmap[y]))
			for x := range heightmap[y] {
				scaledHeightmap[y][x] = (heightmap[y][x] + 1) * MapHeight / 2
			}
		}
		fmt.Printf("\nEscalado del mapa: %.3f segundos\n", time.Since(scaleStart).Seconds())

		imgPath := filepath.Join(imageDir, fmt.Sprintf("terrain_render_%d.png", i))
		meshPath := filepath.Join(meshDir, fmt.Sprintf("eroded_terrain_%d.ply", i))

		meshStart := time.Now()
		vertices, faces, colors := terrain.GenerateHeightmapMesh(scaledHeightmap)
		fmt.Printf("\nGeneración de malla: %.3f segundos\n", time.Since(meshStart).Seconds())

		plyStart := time.Now()
		terrain.SavePLY(meshPath, vertices, faces, colors)
		fmt.Printf("\nGuardado del archivo PLY: %.3f segundos\n", time.Since(plyStart).Seconds())

		renderStart := time.Now()
		terrain.RenderTerrainIsometric(meshPath, imgPath)
		fmt.Printf("\nRenderizado: %.3f segundos\n", time.Since(renderStart).Seconds())

		erosionStart := time.Now()
		heightmap = terrain.ApplyErosion(heightmap, ErosionDropletCount, ErosionParams)
		fmt.Printf("\nAplicación de erosión (%d gotas): %.3f segundos\n", ErosionDropletCount, time.Since(erosionStart).Seconds())

		fmt.Printf("\nTiempo total de iteración %d: %.3f segundos\n", i+1, time.Since(iterStart).Seconds())
	}

	fmt.Printf("\nTiempo total de ejecución: %.3f segundos (%.2f minutos)\n", time.Since(start_time).Seconds(), time.Since(start_time).Minutes())
}

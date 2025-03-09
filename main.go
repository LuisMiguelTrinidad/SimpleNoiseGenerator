package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"

	"main/terrain"

	"github.com/setanarut/rainfall"
)

func createNoiseMap(size int, seed int64, octaves int, scale float64, smoothingFunction func(float64) float64) *terrain.Terrain {
	noise := terrain.NewOpenSimplex(seed)
	terrainObj := terrain.NewTerrain(size)

	// Precompute amplitudeSum, freqs and amps
	freqs := make([]float64, octaves)
	amps := make([]float64, octaves)
	persistence := 0.5
	amplitudeSum := 0.0
	for i := 0; i < octaves; i++ {
		freqs[i] = math.Pow(2, float64(i))
		amps[i] = math.Pow(persistence, float64(i))
		amplitudeSum += amps[i]
	}

	// Generate initial heightmap - Sequential version
	for y := range size {
		for x := range size {
			var total float64
			for i := 0; i < octaves; i++ {
				nx := float64(x) / scale * freqs[i]
				ny := float64(y) / scale * freqs[i]
				total += noise.Eval2(nx, ny) * amps[i]
			}
			value := total / amplitudeSum
			terrainObj.Heightmap[x][y] = smoothingFunction(value) * 256
		}
	}
	return terrainObj
}

// Convierte el terreno a una imagen en escala de grises
func terrainToImage(t *terrain.Terrain) *image.Gray {
	size := len(t.Heightmap)
	img := image.NewGray(image.Rect(0, 0, size, size))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Aseguramos que el valor esté en el rango [0, 255]
			val := int(math.Max(0, math.Min(255, t.Heightmap[x][y])))
			img.SetGray(x, y, color.Gray{uint8(val)})
		}
	}
	return img
}

// Actualiza el terreno con los datos de la imagen erosionada
func imageToTerrain(img image.Image, t *terrain.Terrain) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			// Usar solo el canal rojo (en escala de grises todos son iguales)
			gray := float64(r >> 8)
			t.Heightmap[x][y] = gray
		}
	}
}

// Aplica erosión hidráulica usando la biblioteca rainfall
func applyRainfallErosion(t *terrain.Terrain, dropletCount int,
	friction float64, depositionRate float64, evaporationRate float64) error {

	// Crear directorio temporal si no existe
	tmpDir := "temp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("error al crear directorio temporal: %v", err)
	}

	tempInputFile := filepath.Join(tmpDir, "height_input.png")
	tempOutputFile := filepath.Join(tmpDir, "height_eroded.png")

	// Convertir terreno a imagen y guardarla
	img := terrainToImage(t)
	f, err := os.Create(tempInputFile)
	if err != nil {
		return fmt.Errorf("error al crear archivo temporal: %v", err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		return fmt.Errorf("error al codificar imagen: %v", err)
	}
	f.Close()

	// Configurar opciones de rainfall
	opts := &rainfall.Options{
		Scale:              100.0,           // Escala de altura
		Density:            1,               // Densidad de gotas
		Friction:           friction,        // Fricción/inercia
		DepositionRate:     depositionRate,  // Tasa de deposición
		EvaporationRate:    evaporationRate, // Tasa de evaporación
		RaindropRandomSeed: 42,              // Semilla aleatoria
	}

	// Aplicar erosión
	sim := rainfall.NewFromImageFile(tempInputFile, opts)
	sim.Raindrops(dropletCount)
	sim.WriteToImageFile(tempOutputFile)

	// Cargar resultado y actualizar terreno
	erodedFile, err := os.Open(tempOutputFile)
	if err != nil {
		return fmt.Errorf("error al abrir archivo erosionado: %v", err)
	}

	erodedImg, err := png.Decode(erodedFile)
	if err != nil {
		erodedFile.Close()
		return fmt.Errorf("error al decodificar imagen erosionada: %v", err)
	}
	erodedFile.Close()

	// Actualizar el terreno con los nuevos valores de altura
	imageToTerrain(erodedImg, t)
	return nil
}

func main() {
	start_time := time.Now()

	const (
		size            = 512
		scale           = 2048
		octaves         = 12
		seed            = 0
		dropletCount    = 200000
		friction        = 0.3
		depositionRate  = 0.3
		evaporationRate = 1.0 / 512.0 // Ajustado por tamaño de imagen
	)

	// Create output directory if not exists
	outputDir := "images"
	os.MkdirAll(outputDir, 0755)

	fmt.Printf("%v Generating terrain...\n", time.Since(start_time))

	// Create noise map
	terrainObj := createNoiseMap(size, seed, octaves, scale, func(height float64) float64 {
		for range 1 {
			height = 0.75*math.Copysign((math.Sin(math.Pi*height-math.Pi/2)/2+0.5), height) + 0.25*height
		}
		return height
	})

	fmt.Printf("%v Aplicando erosión con Rainfall...\n", time.Since(start_time))

	// Número de iteraciones de erosión
	const erosionIterations = 10
	dropletsPerIteration := dropletCount / erosionIterations

	for r := range erosionIterations {
		fmt.Printf("%v Iteración de erosión %d/%d...\n", time.Since(start_time), r+1, erosionIterations)

		if err := applyRainfallErosion(terrainObj, dropletsPerIteration,
			friction, depositionRate, evaporationRate); err != nil {
			fmt.Printf("Error en erosión: %v\n", err)
			return
		}

		// Generamos y guardamos el modelo 3D
		vertices, vertexColors := terrainObj.GenerateVertices()
		faces := terrainObj.GenerateFaces()
		terrain.SavePLY(filepath.Join(outputDir, fmt.Sprintf("eroded_terrain%v.ply", r)), vertices, faces, vertexColors)

		// Guardamos la imagen
		img := terrainToImage(terrainObj)
		file, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("eroded_terrain%v.png", r)))
		if err != nil {
			fmt.Println("Error al crear el archivo:", err)
			return
		}
		if err := png.Encode(file, img); err != nil {
			fmt.Println("Error al guardar la imagen:", err)
			file.Close()
			return
		}
		file.Close()
	}

	fmt.Printf("%v Post-procesando terreno...\n", time.Since(start_time))
	fmt.Printf("%v Guardando terreno...\n", time.Since(start_time))
	fmt.Printf("%v Renderizando terreno...\n", time.Since(start_time))
	fmt.Printf("outputDir: %v\n", filepath.Join(outputDir, "eroded_terrain.png"))
	fmt.Printf("%v ¡Hecho!\n", time.Since(start_time))

	// Limpieza de archivos temporales
	os.RemoveAll("temp")
}

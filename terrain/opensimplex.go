package terrain

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// OpenSimplex represents a noise generator with a permutation table and gradients.
type OpenSimplex struct {
	perm [512]int
}

var gradients = [8][2]float64{
	{1.0, 0.0}, {0.7071, 0.7071}, {0.0, 1.0}, {-0.7071, 0.7071},
	{-1.0, 0.0}, {-0.7071, -0.7071}, {0.0, -1.0}, {0.7071, -0.7071},
}

// NewOpenSimplex initializes a new noise generator with a given seed.
func NewOpenSimplex(seed int64) *OpenSimplex {
	os := &OpenSimplex{}
	src := rand.NewSource(seed)
	r := rand.New(src)

	perm := [256]int{}
	for i := range perm {
		perm[i] = i
	}

	// Shuffle the permutation array using the provided seed
	for i := len(perm) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		perm[i], perm[j] = perm[j], perm[i]
	}

	// Duplicate the permutation array to avoid overflow
	for i := 0; i < 512; i++ {
		os.perm[i] = perm[i%256]
	}
	return os
}

// Eval2 computes the 2D noise value at the given coordinates.
func (os *OpenSimplex) Eval2(x, y float64) float64 {
	// ...existing code...
	const (
		F2 = 0.3660254037844386  // (sqrt(3) - 1) / 2
		G2 = 0.21132486540518713 // (3 - sqrt(3)) / 6
	)

	// Skew input to hexagonal grid
	s := (x + y) * F2
	xs := x + s
	ys := y + s
	i := int(math.Floor(xs))
	j := int(math.Floor(ys))

	// ...existing code...
	t := float64(i+j) * G2
	x0 := float64(i) - t
	y0 := float64(j) - t
	x0s := x - x0
	y0s := y - y0

	// Determine which simplex we're in
	var i1, j1 int
	if x0s > y0s {
		i1, j1 = 1, 0 // Lower triangle
	} else {
		i1, j1 = 0, 1 // Upper triangle
	}

	// ...existing code...
	x1 := x0s - float64(i1) + G2
	y1 := y0s - float64(j1) + G2
	x2 := x0s - 1.0 + 2.0*G2
	y2 := y0s - 1.0 + 2.0*G2

	ii := i & 255
	jj := j & 255

	// Calculate gradients from permutation table
	gi0 := os.perm[ii+os.perm[jj]] % 8
	gi1 := os.perm[ii+i1+os.perm[jj+j1]] % 8
	gi2 := os.perm[ii+1+os.perm[jj+1]] % 8

	// ...existing code...
	var n0, n1, n2 float64
	t0 := 0.5 - x0s*x0s - y0s*y0s
	if t0 >= 0 {
		t0 *= t0
		g := gradients[gi0]
		n0 = t0 * t0 * (g[0]*x0s + g[1]*y0s)
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 >= 0 {
		t1 *= t1
		g := gradients[gi1]
		n1 = t1 * t1 * (g[0]*x1 + g[1]*y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 >= 0 {
		t2 *= t2
		g := gradients[gi2]
		n2 = t2 * t2 * (g[0]*x2 + g[1]*y2)
	}

	return 70.0 * (n0 + n1 + n2)
}

func CreateNoiseMap(mapseed int64, mapSize int, mapScale float64, mapOctaves int, smoothingFunction func(float64) float64) [][]float64 {
	startTotal := time.Now()
	fmt.Printf("Iniciando generación de mapa de ruido %dx%d (escala: %.1f, octavas: %d)\n",
		mapSize, mapSize, mapScale, mapOctaves)

	// Inicializar el generador de ruido
	startNoise := time.Now()
	noise := NewOpenSimplex(mapseed)
	fmt.Printf("  ├─ Inicialización del generador: %.3f ms\n",
		float64(time.Since(startNoise).Microseconds())/1000)

	// Crear directamente un array 2D
	startHeightmap := time.Now()
	heightmap := make([][]float64, mapSize)
	for i := range heightmap {
		heightmap[i] = make([]float64, mapSize)
	}
	fmt.Printf("  ├─ Creación de array heightmap: %.3f ms\n",
		float64(time.Since(startHeightmap).Microseconds())/1000)

	// Precompute amplitudeSum, freqs and amps
	startPrecompute := time.Now()
	freqs := make([]float64, mapOctaves)
	amps := make([]float64, mapOctaves)
	persistence := 0.5
	amplitudeSum := 0.0
	for i := range mapOctaves {
		freqs[i] = math.Pow(2, float64(i))
		amps[i] = math.Pow(persistence, float64(i))
		amplitudeSum += amps[i]
	}
	fmt.Printf("  ├─ Precálculo de frecuencias y amplitudes: %.3f ms\n",
		float64(time.Since(startPrecompute).Microseconds())/1000)

	// Generate initial heightmap - Sequential version
	startGeneration := time.Now()
	totalEvals := 0
	for y := 0; y < mapSize; y++ {
		if y > 0 && y%(mapSize/10) == 0 {
			pctComplete := float64(y) / float64(mapSize) * 100
			timeElapsed := time.Since(startGeneration)
			timeEstimated := time.Duration(float64(timeElapsed) / (float64(y) / float64(mapSize)))
			timeRemaining := timeEstimated - timeElapsed
			fmt.Printf("  │  ├─ %.1f%% completado - Tiempo restante: %.1f s\n",
				pctComplete, timeRemaining.Seconds())
		}

		for x := 0; x < mapSize; x++ {
			var total float64
			for i := 0; i < mapOctaves; i++ {
				nx := float64(x) / mapScale * freqs[i]
				ny := float64(y) / mapScale * freqs[i]
				total += noise.Eval2(nx, ny) * amps[i]
				totalEvals++
			}
			heightmap[y][x] = total / amplitudeSum
		}
	}
	generationTime := time.Since(startGeneration)
	fmt.Printf("  ├─ Generación del mapa base: %.3f s (%.1f millones de eval./s)\n",
		generationTime.Seconds(), float64(totalEvals)/(generationTime.Seconds()*1000000))

	// Aplicar función de suavizado
	startSmoothing := time.Now()
	for y := range mapSize {
		for x := range mapSize {
			heightmap[y][x] = smoothingFunction(heightmap[y][x])
		}
	}
	fmt.Printf("  ├─ Aplicación de filtro de suavizado: %.3f ms\n",
		float64(time.Since(startSmoothing).Microseconds())/1000)

	fmt.Printf("  └─ Tiempo total generación de mapa: %.3f s\n", time.Since(startTotal).Seconds())

	return heightmap
}

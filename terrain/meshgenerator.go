package terrain

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"time"
)

// ColorFromHeight genera un color RGB basado en la altura del terreno
func ColorFromHeight(value float64, min, max float64) [3]float64 {
	// Normalizar la altura al rango [0,1]
	normalized := (value - min) / (max - min)
	return [3]float64{
		1.0 - normalized,                   // R: disminuye con la altura
		1.0 - math.Abs(normalized-0.5)*2.0, // G: máximo en el medio
		normalized,                         // B: aumenta con la altura
	}
}

// SavePLY guarda el terreno como un archivo 3D en formato PLY
func SavePLY(filename string, vertices [][3]float64, faces [][3]int, colors [][3]float64) error {
	startTotal := time.Now()
	fmt.Printf("Guardando malla en archivo PLY: %s\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()

	startHeader := time.Now()
	// Escribir cabecera PLY
	fmt.Fprintf(writer, "ply\n")
	fmt.Fprintf(writer, "format ascii 1.0\n")
	fmt.Fprintf(writer, "element vertex %d\n", len(vertices))
	fmt.Fprintf(writer, "property float x\n")
	fmt.Fprintf(writer, "property float y\n")
	fmt.Fprintf(writer, "property float z\n")
	fmt.Fprintf(writer, "property uchar red\n")   // Cambiado a uchar (0-255)
	fmt.Fprintf(writer, "property uchar green\n") // Cambiado a uchar (0-255)
	fmt.Fprintf(writer, "property uchar blue\n")  // Cambiado a uchar (0-255)
	fmt.Fprintf(writer, "element face %d\n", len(faces))
	fmt.Fprintf(writer, "property list uchar int vertex_indices\n")
	fmt.Fprintf(writer, "end_header\n")
	fmt.Printf("  ├─ Escritura de cabecera: %.3f ms\n",
		float64(time.Since(startHeader).Microseconds())/1000)

	// Escribir vértices con colores (convertidos a enteros 0-255)
	startVertices := time.Now()
	for i, v := range vertices {
		c := colors[i]
		// Convertir colores float (0-1) a uchar (0-255)
		r := int(math.Round(c[0] * 255))
		g := int(math.Round(c[1] * 255))
		b := int(math.Round(c[2] * 255))
		fmt.Fprintf(writer, "%.2f %.2f %.2f %d %d %d\n", v[0], v[1], v[2], r, g, b)
	}
	fmt.Printf("  ├─ Escritura de %d vértices: %.3f ms\n",
		len(vertices), float64(time.Since(startVertices).Microseconds())/1000)

	// Escribir caras
	startFaces := time.Now()
	for _, f := range faces {
		fmt.Fprintf(writer, "3 %d %d %d\n", f[0], f[1], f[2])
	}
	fmt.Printf("  ├─ Escritura de %d caras: %.3f ms\n",
		len(faces), float64(time.Since(startFaces).Microseconds())/1000)

	fmt.Printf("  └─ Tiempo total guardado PLY: %.3f ms\n",
		float64(time.Since(startTotal).Microseconds())/1000)

	return nil
}

// GenerateHeightmapMesh crea una malla 3D completa a partir de un heightmap 2D
func GenerateHeightmapMesh(heightmap [][]float64) ([][3]float64, [][3]int, [][3]float64) {
	startTotal := time.Now()
	height := len(heightmap)
	width := len(heightmap[0])
	fmt.Printf("Iniciando generación de malla %dx%d (%d vértices)...\n",
		width, height, width*height)

	// Encontrar valores mínimo y máximo para la coloración
	startMinMax := time.Now()
	minHeight, maxHeight := heightmap[0][0], heightmap[0][0]
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if heightmap[y][x] < minHeight {
				minHeight = heightmap[y][x]
			}
			if heightmap[y][x] > maxHeight {
				maxHeight = heightmap[y][x]
			}
		}
	}
	fmt.Printf("  ├─ Cálculo de rango de alturas: %.3f ms\n",
		float64(time.Since(startMinMax).Microseconds())/1000)
	fmt.Printf("  │  ├─ Altura mínima: %.2f\n", minHeight)
	fmt.Printf("  │  └─ Altura máxima: %.2f\n", maxHeight)

	// Generar vértices y colores
	startVertices := time.Now()
	vertices := make([][3]float64, height*width)
	colors := make([][3]float64, height*width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			h := heightmap[y][x]

			// Posición del vértice
			// 1. Mantenemos x como está
			// 2. Mantenemos y como está (sin invertir)
			// 3. INVERTIMOS el valor de altura (h) para corregir la orientación del eje Z
			vertices[idx] = [3]float64{
				float64(x),
				float64(y),
				-h, // Invertimos el signo de la altura para corregir la orientación en el eje Z
			}

			// Color basado en altura normalizada (usamos h original, no -h, para mantener colores consistentes)
			colors[idx] = ColorFromHeight(h, minHeight, maxHeight)
		}
	}
	fmt.Printf("  ├─ Generación de %d vértices y colores: %.3f ms\n",
		len(vertices), float64(time.Since(startVertices).Microseconds())/1000)

	// Generar caras (triángulos)
	startFaces := time.Now()
	numFaces := 2 * (height - 1) * (width - 1)
	faces := make([][3]int, numFaces)
	faceIdx := 0
	for y := 0; y < height-1; y++ {
		for x := 0; x < width-1; x++ {
			// Índices de los vértices para formar los triángulos
			topLeft := y*width + x
			topRight := y*width + (x + 1)
			bottomLeft := (y+1)*width + x
			bottomRight := (y+1)*width + (x + 1)

			// Como invertimos el eje Z, necesitamos invertir el orden de los vértices
			// para mantener la correcta orientación de las normales
			faces[faceIdx] = [3]int{
				topLeft,
				topRight,
				bottomLeft,
			}
			faceIdx++

			faces[faceIdx] = [3]int{
				bottomLeft,
				topRight,
				bottomRight,
			}
			faceIdx++
		}
	}
	fmt.Printf("  ├─ Generación de %d triángulos: %.3f ms\n",
		numFaces, float64(time.Since(startFaces).Microseconds())/1000)

	// Calcular estadísticas de la malla
	memoryVertices := len(vertices) * 3 * 8 // 3 float64 por vértice (8 bytes cada uno)
	memoryColors := len(colors) * 3 * 8     // 3 float64 por color (8 bytes cada uno)
	memoryFaces := len(faces) * 3 * 4       // 3 int por cara (4 bytes cada uno)
	totalMemory := memoryVertices + memoryColors + memoryFaces

	fmt.Printf("  ├─ Memoria estimada: %.2f MB\n", float64(totalMemory)/(1024*1024))
	fmt.Printf("  │  ├─ Vértices: %.2f MB\n", float64(memoryVertices)/(1024*1024))
	fmt.Printf("  │  ├─ Colores: %.2f MB\n", float64(memoryColors)/(1024*1024))
	fmt.Printf("  │  └─ Caras: %.2f MB\n", float64(memoryFaces)/(1024*1024))

	fmt.Printf("  └─ Tiempo total generación de malla: %.3f ms\n",
		float64(time.Since(startTotal).Microseconds())/1000)

	return vertices, faces, colors
}

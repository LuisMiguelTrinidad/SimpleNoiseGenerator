package terrain

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 1    // optional supersampling
	width  = 1600 // output width in pixels
	height = 1600 // output height in pixels
	fovy   = 30   // vertical field of view, slightly wider than before
	near   = 0.01 // near clipping plane, closer to origin
	far    = 200  // far clipping plane, much further to encompass more space

	// Parámetros para el rango de altura
	minHeightParam = 0   // valor mínimo para el rango de altura
	maxHeightParam = 255 // valor máximo para el rango de altura
)

// ColorStop representa un color a una determinada altura
type ColorStop struct {
	Height float64
	Color  fauxgl.Color
}

// TerrainColorMap define un mapa de colores para diferentes alturas
var TerrainColorMap = []ColorStop{
	{minHeightParam, fauxgl.HexColor("#0077BE")},                                        // Agua profunda
	{minHeightParam + (maxHeightParam-minHeightParam)*0.47, fauxgl.HexColor("#00A9E6")}, // Agua poco profunda
	{minHeightParam + (maxHeightParam-minHeightParam)*0.5, fauxgl.HexColor("#FFD700")},  // Arena/Playa
	{minHeightParam + (maxHeightParam-minHeightParam)*0.51, fauxgl.HexColor("#567D46")}, // Vegetación baja
	{minHeightParam + (maxHeightParam-minHeightParam)*0.52, fauxgl.HexColor("#228B22")}, // Bosque
	{minHeightParam + (maxHeightParam-minHeightParam)*0.53, fauxgl.HexColor("#A0522D")}, // Montaña baja
	{minHeightParam + (maxHeightParam-minHeightParam)*0.56, fauxgl.HexColor("#8B4513")}, // Montaña media
	{maxHeightParam, fauxgl.HexColor("#FFFFFF")},                                        // Nieve/Picos
}

var (
	// Simplified camera position - directly above and closer
	eye    = fauxgl.V(math.Pi, math.Pi, math.Pi)             // Camera position directly above, closer height
	center = fauxgl.V(0, 0, 0)                               // View center at origin
	up     = fauxgl.V(0, 0, 1)                               // Up vector for top-down
	light  = fauxgl.V(math.Pi, math.Pi, math.Pi).Normalize() // Light direction
)

// GetColorForHeight devuelve un color interpolado según la altura
func GetColorForHeight(height float64, minparam, maxparam float64) fauxgl.Color {
	// Convertir de rango [-1, 1] a [minparam, maxparam]
	normalizedHeight := (height+1)*(maxparam-minparam)/2 + minparam

	// Asegurar que la altura está dentro del rango [minparam, maxparam]
	normalizedHeight = math.Max(minparam, math.Min(maxparam, normalizedHeight))

	// Encontrar los límites entre los que está la altura
	for i := 0; i < len(TerrainColorMap)-1; i++ {
		if normalizedHeight >= TerrainColorMap[i].Height && normalizedHeight <= TerrainColorMap[i+1].Height {
			return TerrainColorMap[i].Color
		}
	}

	// Si la altura es menor que el primer punto, devolver el primer color
	if normalizedHeight < TerrainColorMap[0].Height {
		return TerrainColorMap[0].Color
	}

	// Si la altura es mayor que el último punto o no se encontró un rango,
	// devolver el último color
	return TerrainColorMap[len(TerrainColorMap)-1].Color
}

// RenderTerrainIsometric function renders a .ply terrain file in isometric view
func RenderTerrainIsometric(plyFilePath string, outputFilePath string) {
	startTotal := time.Now()
	fmt.Printf("Iniciando renderizado del terreno en vista isométrica...\n")

	// Load the mesh from the PLY file
	startLoad := time.Now()
	mesh, err := fauxgl.LoadPLY(plyFilePath)
	if err != nil {
		log.Fatalf("Error loading PLY file: %v", err)
		return // Exit function on error
	}
	fmt.Printf("  ├─ Carga del archivo PLY: %.3f ms\n",
		float64(time.Since(startLoad).Microseconds())/1000)
	fmt.Printf("  │  ├─ Triángulos: %d\n", len(mesh.Triangles))
	fmt.Printf("  │  └─ Vértices aproximados: %d\n", len(mesh.Triangles)*3)

	// Fit mesh in a bi-unit cube centered at the origin
	startNormalize := time.Now()
	mesh.BiUnitCube()
	fmt.Printf("  ├─ Normalización del modelo: %.3f ms\n",
		float64(time.Since(startNormalize).Microseconds())/1000)

	// Aplicar colores basados en la altura a cada vértice
	startColoring := time.Now()
	for i := range mesh.Triangles {
		t := mesh.Triangles[i] // Usar el puntero directamente

		height1 := t.V1.Position.Z
		t.V1.Color = GetColorForHeight(height1, minHeightParam, maxHeightParam)

		height2 := t.V2.Position.Z
		t.V2.Color = GetColorForHeight(height2, minHeightParam, maxHeightParam)

		height3 := t.V3.Position.Z
		t.V3.Color = GetColorForHeight(height3, minHeightParam, maxHeightParam)
	}
	fmt.Printf("  ├─ Aplicación de colores: %.3f ms\n",
		float64(time.Since(startColoring).Microseconds())/1000)

	// Smoothing enabled
	startSmoothing := time.Now()
	mesh.SmoothNormalsThreshold(fauxgl.Radians(30))
	fmt.Printf("  ├─ Suavizado de normales: %.3f ms\n",
		float64(time.Since(startSmoothing).Microseconds())/1000)

	// Create a rendering context
	startContext := time.Now()
	context := fauxgl.NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(fauxgl.HexColor("#00000000"))
	fmt.Printf("  ├─ Creación del contexto de renderizado: %.3f ms\n",
		float64(time.Since(startContext).Microseconds())/1000)

	// Create transformation matrix for isometric view
	startMatrix := time.Now()
	aspect := float64(width) / float64(height)
	// Perspective matrix
	matrix := fauxgl.LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// Usar shader que respeta los colores de los vértices
	shader := fauxgl.NewPhongShader(matrix, light, eye)
	context.Shader = shader
	fmt.Printf("  ├─ Configuración de matriz y shader: %.3f ms\n",
		float64(time.Since(startMatrix).Microseconds())/1000)

	// Render the mesh
	startRender := time.Now()
	context.DrawMesh(mesh)
	fmt.Printf("  ├─ Renderizado del modelo: %.3f ms\n",
		float64(time.Since(startRender).Microseconds())/1000)

	// Downsample image for antialiasing
	startDownsample := time.Now()
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	fmt.Printf("  ├─ Redimensionado de imagen: %.3f ms\n",
		float64(time.Since(startDownsample).Microseconds())/1000)

	// Save the image to the specified output file
	startSave := time.Now()
	err = fauxgl.SavePNG(outputFilePath, image)
	if err != nil {
		log.Fatalf("Error saving PNG file: %v", err)
	}
	fmt.Printf("  ├─ Guardado de imagen PNG: %.3f ms\n",
		float64(time.Since(startSave).Microseconds())/1000)

	fmt.Printf("  └─ Tiempo total de renderizado: %.3f s\n",
		time.Since(startTotal).Seconds())
}

package terrain

import (
	"bufio"
	"fmt"
	"image/color"
	"math"
	"os"
)

// ColorFromValue generates a color based on the terrain height value.
func ColorFromValue(value float64) color.RGBA {
	normalized := (value + 1) / 2 // Map to [0,1]
	return color.RGBA{
		R: uint8(255 * (1 - normalized)),                 // Red decreases with height
		G: uint8(255 * (1 - math.Abs(normalized-0.5)*2)), // Green peaks at mid
		B: uint8(255 * normalized),                       // Blue increases with height
		A: 255,
	}
}

// SavePLY saves the terrain as a PLY format 3D model file.
func SavePLY(filename string, vertices, faces []string, colors []color.RGBA) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write PLY header
	fmt.Fprintf(writer, "ply\n")
	fmt.Fprintf(writer, "format ascii 1.0\n")
	fmt.Fprintf(writer, "element vertex %d\n", len(vertices))
	fmt.Fprintf(writer, "property float x\n")
	fmt.Fprintf(writer, "property float y\n")
	fmt.Fprintf(writer, "property float z\n")
	fmt.Fprintf(writer, "property uchar red\n")
	fmt.Fprintf(writer, "property uchar green\n")
	fmt.Fprintf(writer, "property uchar blue\n")
	fmt.Fprintf(writer, "element face %d\n", len(faces))
	fmt.Fprintf(writer, "property list uchar int vertex_indices\n")
	fmt.Fprintf(writer, "end_header\n")

	// Write vertices with colors
	for i, v := range vertices {
		col := colors[i]
		fmt.Fprintf(writer, "%s %d %d %d\n", v, col.R, col.G, col.B)
	}

	// Write faces
	for _, f := range faces {
		fmt.Fprintln(writer, f)
	}
}

// GenerateVertices creates vertices from the terrain heightmap.
func (t *Terrain) GenerateVertices() ([]string, []color.RGBA) {
	vertices := make([]string, 0, t.Size*t.Size)
	colors := make([]color.RGBA, 0, t.Size*t.Size)

	for y := range t.Size {
		for x := range t.Size {
			value := t.Heightmap[x][y]/(256) - 1 // Reverse scaling
			vertices = append(vertices, fmt.Sprintf("%d %d %.2f", x, y, t.Heightmap[x][y]))
			colors = append(colors, ColorFromValue(value))
		}
	}
	return vertices, colors
}

// GenerateFaces creates triangular faces for the terrain mesh.
func (t *Terrain) GenerateFaces() []string {
	faces := make([]string, 0, (t.Size-1)*(t.Size-1)*2)
	for y := range t.Size - 1 {
		for x := range t.Size - 1 {
			idx := y*t.Size + x
			faces = append(faces,
				fmt.Sprintf("3 %d %d %d", idx, idx+1, idx+t.Size),
				fmt.Sprintf("3 %d %d %d", idx+1, idx+t.Size+1, idx+t.Size),
			)
		}
	}
	return faces
}

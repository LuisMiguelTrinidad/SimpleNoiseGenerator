package terrain

import (
	"math"
	"math/rand"
)

// Terrain represents a heightmap terrain.
type Terrain struct {
	Heightmap [][]float64
	Size      int
}

// NewTerrain creates a new terrain with the given size.
func NewTerrain(size int) *Terrain {
	t := &Terrain{
		Size:      size,
		Heightmap: make([][]float64, size),
	}
	for i := range t.Heightmap {
		t.Heightmap[i] = make([]float64, size)
	}
	return t
}

// ErosionParams contains parameters for hydraulic erosion simulation.
type ErosionParams struct {
	Iterations       int
	DropletCount     int
	Inertia          float64
	SedimentCapacity float64
	Evaporation      float64
	Gravity          float64
	Seed             int64
}

// ApplyErosion applies hydraulic erosion to the terrain using the specified parameters.
func (t *Terrain) ApplyErosion(params ErosionParams) {
	rand.New(rand.NewSource(params.Seed))

	for range params.DropletCount {
		// Initialize droplet
		x := rand.Float64() * float64(t.Size-1)
		y := rand.Float64() * float64(t.Size-1)
		dirX := 0.0
		dirY := 0.0
		speed := 1.0
		water := 1.0
		sediment := 0.0

		for range params.Iterations {
			// Add NaN checks
			if math.IsNaN(x) || math.IsNaN(y) {
				break
			}

			// Convert coordinates to grid indices
			xi := int(x)
			yi := int(y)
			if xi < 0 || xi >= t.Size-1 || yi < 0 || yi >= t.Size-1 {
				break
			}

			height := t.Heightmap[xi][yi]
			gradX := t.Heightmap[xi+1][yi] - height
			gradY := t.Heightmap[xi][yi+1] - height

			dirX = dirX*params.Inertia - gradX*(1-params.Inertia)
			dirY = dirY*params.Inertia - gradY*(1-params.Inertia)

			lenDir := math.Sqrt(dirX*dirX + dirY*dirY)
			if math.IsNaN(lenDir) {
				break
			}

			if lenDir > 0 {
				dirX /= lenDir
				dirY /= lenDir
			}

			newX := x + dirX
			newY := y + dirY
			if newX < 0 || newX >= float64(t.Size) || newY < 0 || newY >= float64(t.Size) {
				break
			}
			x, y = newX, newY

			xiNew := int(x)
			yiNew := int(y)
			if xiNew < 0 || xiNew >= t.Size || yiNew < 0 || yiNew >= t.Size {
				break
			}

			deltaHeight := t.Heightmap[xiNew][yiNew] - height
			speedSquared := speed*speed + deltaHeight*params.Gravity
			speed = math.Sqrt(math.Max(0, speedSquared))

			capacity := math.Max(-deltaHeight*speed*water*params.SedimentCapacity, 0.01)

			if sediment > capacity || deltaHeight > 0 {
				deposit := (sediment - capacity) * 0.1
				if deltaHeight > 0 {
					deposit = math.Min(deltaHeight, sediment)
				}
				sediment -= deposit
				t.Heightmap[xi][yi] += deposit
			} else {
				erosion := math.Min((capacity-sediment)*0.1, -deltaHeight)
				sediment += erosion
				t.Heightmap[xi][yi] -= erosion
			}

			water *= 1 - params.Evaporation
		}
	}
}

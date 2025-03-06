package terrain

import (
	"math"
	"math/rand"
)

// OpenSimplex represents a noise generator with a permutation table and gradients.
type OpenSimplex struct {
	perm [512]int
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

var gradients = [8][2]float64{
	{1.0, 0.0}, {0.7071, 0.7071}, {0.0, 1.0}, {-0.7071, 0.7071},
	{-1.0, 0.0}, {-0.7071, -0.7071}, {0.0, -1.0}, {0.7071, -0.7071},
}

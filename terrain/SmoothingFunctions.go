package terrain

import (
	"math"
)

func GreatPlains(height float64) float64 {
	return math.Copysign((math.Sin(math.Pi*height-math.Pi/2)/2 + 0.5), height)
}

func Cliff(height float64) float64 {
	return math.Copysign(math.Sqrt(math.Abs(height)), height)
}

func Plateau(height float64, level float64) float64 {
	return level*GreatPlains(height) +
		(1-level)*math.Copysign(math.Sqrt(math.Abs(height)), height)
}

func Molone(height float64, level float64) float64 {
	return (1-height)*(level*Cliff(height)) +
		(height)*((1-level)*GreatPlains(height))
}

package utils

import "fmt"

var (
	units = []string{"B", "Kb", "Mb", "Gb", "Tb", "Pb"}
)

func HumanReadableSize(n int) string {
	x := float32(n)
	i := 0
	for x >= 1024 {
		x /= 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", x, units[i])
}

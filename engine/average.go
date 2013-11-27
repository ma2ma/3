package engine

import (
	"github.com/mumax/3/cuda"
	"github.com/mumax/3/data"
)

func init() {
	DeclFunc("average", Average, "Average of space-dependent quantity over magnet volume")
}

// Take average of quantity, taking into account the volume in which it is defined. E.g.:
// 	Average(m)           // averges only inside magnet.
// 	Average(m.Region(1)) // averages only in region 1
func Average(s Slicer) []float64 {
	buf, recycle := s.Slice()
	if recycle {
		defer cuda.Recycle(buf)
	}
	if v, ok := s.(volumer); ok {
		avg := averageVolume(buf, nil)
		spacefill := v.volume()
		for i := range avg {
			avg[i] /= spacefill
		}
		return avg
	} else {
		return averageVolume(buf, vol())
	}
}

// Average of quantity with explicitly given volume mask.
func averageVolume(b, vol *data.Slice) []float64 {
	nComp := b.NComp()
	nCell := float64(b.Len())
	avg := make([]float64, nComp)
	for i := range avg {
		if vol.IsNil() {
			avg[i] = float64(cuda.Sum(b.Comp(i))) / nCell
		} else {
			avg[i] = float64(cuda.Dot(b.Comp(i), vol)) / (spaceFill() * nCell)
		}
	}
	return avg
}

type volumer interface {
	volume() float64 // normalized volume where quantity is defined (0..1)
}

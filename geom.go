package qrst

import (
	"fmt"
)

const (
	Disk320K  = 1
	Disk1200K = 2
	Disk720K  = 3
	Disk1440K = 4
)

// Geometry stores disk geometry as far as the QRST format cares about
// it.
type Geometry struct {
	Heads           int
	Cylinders       int
	SectorSize      int
	SectorsPerTrack int
	DiskSize        int
}

// TrackLength returns the size of data tracks on this disk.
func (g Geometry) TrackLength() int {
	return g.SectorSize * g.SectorsPerTrack
}

// TrackOffset computes the start of a track given a head and
// cylinder.
func (g Geometry) TrackOffset(head int, cyl int) (int, error) {
	if head < 0 || head > g.Heads {
		return 0, fmt.Errorf("qrst: head falls outside disk geometry (%d/%d)",
			head, g.Heads)
	}
	if cyl < 0 || cyl > g.Cylinders {
		return 0, fmt.Errorf("qrst: cylinder falls outside disk geometry (%d/%d)",
			cyl, g.Cylinders)
	}

	offset := (cyl*g.Heads + head) * (g.SectorsPerTrack * g.SectorSize)
	if offset < 0 {
		return 0, fmt.Errorf("qrst: negative track offset")
	}
	return offset, nil
}

var GeometryFromCapacity = map[byte]Geometry{
	Disk320K:  Geometry{1, 80, 512, 9, 360 * 1024},   // 360K
	Disk720K:  Geometry{2, 79, 512, 9, 720 * 1024},   // 720K
	Disk1440K: Geometry{2, 79, 512, 18, 1440 * 1024}, // 1.44M
}

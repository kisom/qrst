package qrst

import "testing"

func TestGeom(t *testing.T) {
	var offsets = map[int]bool{}
	geom := GeometryFromCapacity[4]

	for cylinder := 0; cylinder < geom.Cylinders; cylinder++ {
		for head := 0; head < geom.Heads; head++ {
			t.Logf("C:%02d, H:%d", cylinder, head)
			offset, err := geom.TrackOffset(head, cylinder)
			if err != nil {
				t.Fatal(err)
			}

			if _, ok := offsets[offset]; ok {
				t.Logf("cyl=%d, head=%d", cylinder, head)
				t.Fatalf("qrst: offset %d already seen", offset)
			}

			offsets[offset] = true
		}
	}
}

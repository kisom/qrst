// The QRST (Quick Release Sector Transfer) disc image format was used
// by Compaq to distribute disk images of diagnostic software. The
// file QRST.EXE or QRST5.EXE would be supplied with the disc images
// to write them to a floppy drive.
package qrst

// Made possible by http://fileformats.archiveteam.org/wiki/Quick_Release_Sector_Transfer.

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// HeaderLength stores the fixed QRST header length.
const HeaderLength = 796

// TODO: convert these to new Go errors (idgaf rn)
var (
	ErrInvalidHeader = errors.New("qrst: header is too short")
	ErrBadMagic      = errors.New("qrst: bad magic in header")
	ErrBadTrailer    = errors.New("qrst: bad trailer in header")
	ErrBadDataHeader = errors.New("qrst: bad data record header")
)

var magic = []byte("QRST")

// "The compressed data stream consists of alternating literal runs (a
// byte giving the length of the run, followed by that number of bytes
// data) and compressed runs (two bytes; first gives number of
// repeats, second gives byte to repeat)."
func decompress(cd []byte, tracklen int) (dat []byte, err error) {
	for {
		if len(cd) == 0 {
			break
		}
		runLength := int(cd[0])
		cd = cd[1:]
		dat = append(dat, cd[:runLength]...)
		cd = cd[runLength:]

		if len(cd) == 0 {
			break
		}

		runLength = int(cd[0])
		repeated := cd[1]
		cd = cd[2:]
		for i := runLength; i > 0; i-- {
			dat = append(dat, repeated)
		}
	}

	if tracklen != len(dat) {
		return nil, fmt.Errorf("qrst: decompressed data isn't a track length (%d != %d)", tracklen, len(dat))
	}

	return
}

// CapacityToString returns an appropriate string for a given
// capacity.
func CapacityToString(cap byte) string {
	switch cap {
	case 0:
		return "unknown"
	case 1:
		return "360K"
	case 2:
		return "1.2M"
	case 3:
		return "720K"
	case 4:
		return "1.4M"
	case 5:
		return "160K"
	case 6:
		return "180K"
	case 7:
		return "320K"
	default:
		panic("invalid capacity")
	}
}

// Header describes the QRST image.
type Header struct {
	Raw           []byte
	Magic         [4]byte
	Version       float32
	Checksum      [4]byte
	DiskCapacity  byte
	CurrentVolume byte
	VolumeCount   byte
	Description   [96]byte
	DiskLabel     [720]byte
	Trailer       byte

	Geometry Geometry
}

func readHeader(r io.Reader, h *Header) error {
	h.Raw = make([]byte, HeaderLength)
	n, err := r.Read(h.Raw)
	if n != HeaderLength {
		return ErrInvalidHeader
	} else if err != nil {
		// Note: io.EOF is also an actual error, because
		// we should have header + data.
		return err
	}

	copy(h.Magic[:], h.Raw[:4])
	if !bytes.Equal(h.Magic[:], magic) {
		return ErrBadMagic
	}

	bits := binary.LittleEndian.Uint32(h.Raw[4:8])
	h.Version = math.Float32frombits(bits)
	copy(h.Checksum[:], h.Raw[8:12])

	h.DiskCapacity = h.Raw[12]

	switch h.DiskCapacity {
	case 3: // 720K
		h.Geometry = Geometry{1, 79, 512, 9, 720 * 1024}
	case 4:
		h.Geometry = Geometry{1, 79, 512, 18, 1440 * 1024}
	}

	h.CurrentVolume = h.Raw[13]
	h.VolumeCount = h.Raw[14]
	copy(h.Description[:], h.Raw[15:])
	copy(h.DiskLabel[:], h.Raw[0x4B:])
	h.Trailer = h.Raw[0x031B]

	if h.Version < 5.0 {
		if h.Trailer != 0 {
			return ErrBadTrailer
		}
	} else if h.Version >= 5.0 {
		if h.Trailer != 2 {
			return ErrBadTrailer
		}
	}

	// NB: I wrote this trying to read a version 1.0 file, so I'm
	// not going to try parsing the V5+ extra headers --- I don't
	// have a file to use, anyhow.

	return nil
}

// Data represents a track's worth of data in the image.
type Data struct {
	Cylinder int
	Head     int
	Type     byte
	Length   uint16
	Data     []byte
}

// Files have a header and some data records.
type File struct {
	Header Header
	Data   []Data
}

func readNextData(r io.Reader, tracklen int) (Data, error) {
	dat := Data{}

	var header [3]byte

	n, err := r.Read(header[:])
	if err == io.EOF {
		return dat, err
	} else if n != 3 {
		return dat, ErrBadDataHeader
	} else if err != nil {
		// io.EOF is an error because there should be data too
		return dat, err
	}

	dat.Cylinder = int(header[0])
	dat.Head = int(header[1])
	dat.Type = header[2]

	switch dat.Type {
	case 0: // uncompressed track
		dat.Data = make([]byte, tracklen)
		n, err = r.Read(dat.Data)
		if n != tracklen {
			return dat, errors.New("*** short read ***")
		} else if err != nil {
			return dat, err
		}
	case 1: // blank track
		dat.Data = make([]byte, 1)
		n, err = r.Read(dat.Data)
		if n != 1 {
			return dat, errors.New("*** bad filler byte ***")
		}
		if err != nil {
			return dat, err
		}
		dat.Data = make([]byte, tracklen)
	case 2: // compressed track
		err = binary.Read(r, binary.LittleEndian, &dat.Length)
		if err != nil {
			return dat, err
		}
		li := int(dat.Length)
		dat.Data = make([]byte, li)
		n, err = r.Read(dat.Data)
		if n != li {
			return dat, errors.New("*** not enough data ***")
		} else if err != nil {
			return dat, err
		}

		dat.Data, err = decompress(dat.Data, tracklen)
		if err != nil {
			return dat, err
		}
	default:
		return dat, errors.New("*** invalid type ***")
	}

	return dat, nil
}

// LoadFile reads the file from an io Reader; it doesn't assemble the
// data.
func LoadFile(r io.Reader) (*File, error) {
	file := &File{}
	err := readHeader(r, &file.Header)
	if err != nil {
		return nil, err
	}

	for {
		var dat Data
		dat, err = readNextData(r, file.Header.Geometry.TrackLength())
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		file.Data = append(file.Data, dat)
	}

	return file, err
}

// Assemble builds a disk image from a file.
func (f *File) Assemble() ([]byte, error) {
	buffer := make([]byte, f.Header.Geometry.DiskSize)

	for _, data := range f.Data {
		offset, err := f.Header.Geometry.TrackOffset(data.Head, data.Cylinder)
		if err != nil {
			return nil, err
		}
		copy(buffer[offset:], data.Data)
	}
	return buffer, nil
}

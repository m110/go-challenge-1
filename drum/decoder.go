package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

const trackSteps = 16

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  string
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is the representation of single track in a pattern
type Track struct {
	ID    byte
	Name  string
	Steps []byte
}

func (p *Pattern) String() string {
	var result []string

	result = append(result, fmt.Sprintf("Saved with HW Version: %s", p.Version))
	result = append(result, fmt.Sprintf("Tempo: %v", p.Tempo))

	for _, track := range p.Tracks {
		line := fmt.Sprintf("(%d) %s\t", track.ID, track.Name)

		for i, step := range track.Steps {
			if i%4 == 0 {
				line += "|"
			}

			if step == 1 {
				line += "x"
			} else {
				line += "-"
			}

		}

		line += "|"

		result = append(result, line)
	}

	return strings.Join(result, "\n") + "\n"
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return decode(file)
}

type decoder struct {
	file    *os.File
	lastErr error
	pattern *Pattern
}

func decode(file *os.File) (*Pattern, error) {
	d := &decoder{
		file:    file,
		pattern: &Pattern{},
	}

	d.checkHeader()

	length := d.readLength()
	maxOffset := d.currentOffset() + length

	d.readVersion()
	d.readTempo()

	for d.currentOffset() < maxOffset {
		d.readTrack()
	}

	if d.lastErr != nil {
		return nil, d.lastErr
	}

	return d.pattern, nil
}

func (d *decoder) currentOffset() uint64 {
	offset, err := d.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		d.lastErr = err
	}

	return uint64(offset)
}

func (d *decoder) read(v interface{}) {
	var err error

	switch v.(type) {
	case *float32, *float64, *[]float32, *[]float64:
		err = binary.Read(d.file, binary.LittleEndian, v)
	default:
		err = binary.Read(d.file, binary.BigEndian, v)
	}

	if err != nil {
		d.lastErr = err
	}
}

func (d *decoder) checkHeader() {
	if d.lastErr != nil {
		return
	}

	var header = make([]byte, 6)
	d.read(header)

	if string(header) != "SPLICE" {
		d.lastErr = errors.New("Invalid header")
	}
}

func (d *decoder) readLength() uint64 {
	if d.lastErr != nil {
		return 0
	}

	var length uint64
	d.read(&length)

	return length
}

func (d *decoder) readVersion() {
	if d.lastErr != nil {
		return
	}

	var version = make([]byte, 32)
	d.read(version)
	version = bytes.Trim(version, "\x00")

	d.pattern.Version = string(version)
}

func (d *decoder) readTempo() {
	if d.lastErr != nil {
		return
	}

	d.read(&d.pattern.Tempo)
}

func (d *decoder) readTrack() {
	if d.lastErr != nil {
		return
	}

	track := &Track{}

	d.read(&track.ID)

	var length uint32
	d.read(&length)

	var name = make([]byte, length)
	d.read(name)
	name = bytes.Trim(name, "\x00")

	track.Name = string(name)

	var steps = make([]byte, trackSteps)
	d.read(steps)
	track.Steps = steps

	d.pattern.Tracks = append(d.pattern.Tracks, track)
}

package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	Tracks  []Track

	lastErr error
	buffer  io.ReadSeeker
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
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}
	err = p.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Pattern) UnmarshalBinary(data []byte) error {
	p.buffer = bytes.NewReader(data)

	p.checkHeader()

	length := p.readLength()
	maxOffset := p.currentOffset() + length

	p.readVersion()
	p.readTempo()

	for p.currentOffset() < maxOffset {
		p.readTrack()
	}

	if p.lastErr != nil {
		return p.lastErr
	}

	return nil
}

func (p *Pattern) currentOffset() uint64 {
	offset, err := p.buffer.Seek(0, os.SEEK_CUR)
	if err != nil {
		p.lastErr = err
	}

	return uint64(offset)
}

func (p *Pattern) read(v interface{}) {
	var err error

	switch v.(type) {
	case *float32, *float64, *[]float32, *[]float64:
		err = binary.Read(p.buffer, binary.LittleEndian, v)
	default:
		err = binary.Read(p.buffer, binary.BigEndian, v)
	}

	if err != nil {
		p.lastErr = err
	}
}

func (p *Pattern) checkHeader() {
	if p.lastErr != nil {
		return
	}

	var header = make([]byte, 6)
	p.read(header)

	if string(header) != "SPLICE" {
		p.lastErr = errors.New("Invalid header")
	}
}

func (p *Pattern) readLength() uint64 {
	if p.lastErr != nil {
		return 0
	}

	var length uint64
	p.read(&length)

	return length
}

func (p *Pattern) readVersion() {
	if p.lastErr != nil {
		return
	}

	var version = make([]byte, 32)
	p.read(version)
	version = bytes.Trim(version, "\x00")

	p.Version = string(version)
}

func (p *Pattern) readTempo() {
	if p.lastErr != nil {
		return
	}

	p.read(&p.Tempo)
}

func (p *Pattern) readTrack() {
	if p.lastErr != nil {
		return
	}

	track := Track{}

	p.read(&track.ID)

	var length uint32
	p.read(&length)

	var name = make([]byte, length)
	p.read(name)
	name = bytes.Trim(name, "\x00")

	track.Name = string(name)

	var steps = make([]byte, trackSteps)
	p.read(steps)
	track.Steps = steps

	p.Tracks = append(p.Tracks, track)
}

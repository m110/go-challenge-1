package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  string
	Version string
	Tempo   float64
	Tracks  []struct {
		Id    uint32
		Name  string
		Steps [16]byte
	}
}

func (p *Pattern) String() string {
	var result []string

	result = append(result, fmt.Sprintf("Saved with HW Version: %s", p.Version))
	result = append(result, fmt.Sprintf("Tempo: %f", p.Tempo))

	for _, track := range p.Tracks {
		result = append(result, fmt.Sprintf("(%d) %s", track.Id, track.Name))
	}

	return strings.Join(result, "\n")
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

	return Decode(file)
}

type decoder struct {
	reader  io.Reader
	lastErr error
	pattern *Pattern

	length uint64
}

func Decode(reader io.Reader) (*Pattern, error) {
	d := &decoder{
		reader:  reader,
		pattern: &Pattern{},
	}

	d.checkHeader()
	d.decodeLength()

	if d.lastErr != nil {
		return nil, d.lastErr
	}

	return d.pattern, nil
}

func (d *decoder) read(v interface{}) {
	err := binary.Read(d.reader, binary.BigEndian, v)
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

	fmt.Println("DEBUG Header:", string(header))

	if string(header) != "SPLICE" {
		d.lastErr = errors.New("Invalid header")
	}
}

func (d *decoder) decodeLength() {
	if d.lastErr != nil {
		return
	}

	d.read(&d.length)

	fmt.Println("DEBUG length:", d.length)
}

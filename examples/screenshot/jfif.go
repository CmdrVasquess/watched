package screenshot

import (
	"bufio"
	"fmt"
	"io"
)

type JFIFMarkerTag byte

func (t JFIFMarkerTag) String() string {
	spec := &jfifMSpecs[t]
	if spec.name == "" {
		return fmt.Sprintf("<?> (%02X)", byte(t))
	}
	return fmt.Sprintf("%s (%02X)", spec.name, byte(t))
}

func (t JFIFMarkerTag) Segment() bool {
	spec := &jfifMSpecs[t]
	return spec.segment
}

func (t JFIFMarkerTag) WriteMarker(w io.Writer, size uint16) error {
	marker := [4]byte{0xFF, byte(t), byte(size >> 8), byte(size)}
	if t.Segment() {
		_, err := w.Write(marker[:])
		return err
	}
	_, err := w.Write(marker[:2])
	return err
}

const (
	JFIFMarkerSOI   JFIFMarkerTag = 0xD8
	JFIFMarkerAPP0  JFIFMarkerTag = 0xE0 // JFIF Tag
	JFIFMarkerDAC   JFIFMarkerTag = 0xCC
	JFIFMarkerDQT   JFIFMarkerTag = 0xDB
	JFIFMarkerDRI   JFIFMarkerTag = 0xDD
	JFIFMarkerAPP1  JFIFMarkerTag = 0xE1 // EXIF Tag
	JFIFMarkerAPP14 JFIFMarkerTag = 0xEE // Often for copyright
	JFIFMarkerCOM   JFIFMarkerTag = 0xFE
	JFIFMarkerSOS   JFIFMarkerTag = 0xDA
	JFIFMarkerEOI   JFIFMarkerTag = 0xD9
)

// JFIFMarkerC does not check the range of n
func JFIFMarkerSOF(n int) JFIFMarkerTag { return 0xC0 | (JFIFMarkerTag(n) & 0xf) }

func JFIFMarkerAPP(n int) JFIFMarkerTag { return 0xE0 | (JFIFMarkerTag(n) & 0xf) }

type JFIFScanner struct {
	rd   io.Reader
	Err  error
	Tag  JFIFMarkerTag
	Size int
}

func NewJFIFScanner(r io.Reader) *JFIFScanner {
	return &JFIFScanner{rd: r}
}

func (s *JFIFScanner) Scan() bool {
	var marker [2]byte
	if n, err := s.rd.Read(marker[:]); err != nil {
		s.Err = err
		return false
	} else if n < len(marker) {
		s.Err = fmt.Errorf("incomplete marker %v", marker[:n])
		return false
	}
	if marker[0] != 0xFF {
		s.Err = fmt.Errorf("marker does not start with 0xFF: %v", marker)
		return false
	}
	s.Tag = JFIFMarkerTag(marker[1])
	spec := &jfifMSpecs[s.Tag]
	if !spec.segment {
		s.Size = 0
		return true
	}
	if spec.name == "" {
		s.Err = fmt.Errorf("unknown marker tag %02X", s.Tag)
		return false
	}
	if n, err := s.rd.Read(marker[:]); err != nil {
		s.Err = err
		return false
	} else if n < len(marker) {
		s.Err = fmt.Errorf("incomplete segment size %v", marker[:n])
		return false
	}
	s.Size = 256*int(marker[0]) + int(marker[1])
	return true
}

func (s *JFIFScanner) Segment() io.Reader {
	switch {
	case s.Tag == JFIFMarkerSOS:
		return (*sosReader)(bufio.NewReader(s.rd))
	case s.Size < 2:
		return &io.LimitedReader{}
	}
	return &io.LimitedReader{
		R: s.rd,
		N: int64(s.Size - 2),
	}
}

func (s *JFIFScanner) Consume(segment io.Reader) (int64, error) {
	n, err := io.Copy(io.Discard, segment)
	return n, err
}

type jfifMarcerSpec struct {
	name    string
	segment bool
}

var jfifMSpecs [256]jfifMarcerSpec

func init() {
	for i := 0; i < 16; i++ {
		jfifMSpecs[JFIFMarkerAPP(i)] = jfifMarcerSpec{
			fmt.Sprintf("APP%d", i),
			true,
		}
		jfifMSpecs[JFIFMarkerSOF(i)] = jfifMarcerSpec{
			fmt.Sprintf("SOF%d", i),
			true,
		}
	}
	jfifMSpecs[JFIFMarkerSOI] = jfifMarcerSpec{"SOI", false}
	jfifMSpecs[JFIFMarkerDAC] = jfifMarcerSpec{"DAC", true}
	jfifMSpecs[JFIFMarkerDQT] = jfifMarcerSpec{"DQT", true}
	jfifMSpecs[JFIFMarkerDRI] = jfifMarcerSpec{"DRI", true}
	jfifMSpecs[JFIFMarkerCOM] = jfifMarcerSpec{"COM", true}
	jfifMSpecs[JFIFMarkerSOS] = jfifMarcerSpec{"SOS", false}
	jfifMSpecs[JFIFMarkerEOI] = jfifMarcerSpec{"EOI", false}
}

type sosReader bufio.Reader

func (r *sosReader) Read(p []byte) (n int, err error) {
	b := (*bufio.Reader)(r)
	for n = 0; n < len(p); n++ {
		q, err := b.ReadByte()
		if err != nil {
			return n, err
		}
		if q != 0xff {
			p[n] = q
			continue
		}
		if ahead, err := b.Peek(1); err != nil {
			p[n] = q
			return n + 1, err
		} else if JFIFMarkerTag(ahead[0]) == JFIFMarkerEOI {
			return n, io.EOF
		}
		p[n] = q
		b.UnreadByte()
	}
	return n, nil
}

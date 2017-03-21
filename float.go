package listreader

import "io"
import "math"
import "bytes"
import "errors"
//import "fmt"

type progress uint8

const (
	begin progress = iota
	start
	inWhole
	beginFraction
	inFraction
	exponentSign
	inExponent
	nan
)

const maxUint = math.MaxUint64 / 10

// Floats is the state, and behaviour, of a floating-point Reader.
type Floats struct {
	io.Reader
	Delimiter      byte
	stage          progress
	neg            bool   // negative number
	whole          uint64 // whole number section, read so far
	fraction       uint64 // fraction section read so far
	fractionDigits uint8  // count of fractional section digits, used to turn int into required real number by power of ten division
	exponent       uint64 // exponent section so far read
	negExponent    bool
	AnyNaN         bool   // set if any parsing issue
	buf            []byte // reusable buffer for reader, same backing array as Unbuf.
	UnBuf          []byte // unconsumed bytes remaining in internal buffer after last read
	LastDelimiter  byte   // delimiter after last item
}

// NewFloats returns a Floats reading items from r separated by d, with the default buffer size.
func NewFloats(r io.Reader, d byte) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bytes.MinRead)}
}

// NewFloatsSize returns a Floats reading items from r separated by d, with an internal buffer size of bSize.
func NewFloatsSize(r io.Reader, d byte, bSize int) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bSize)}
}

type ErrAnyNaN struct {
	error
}

// ReadAll returns all the floating-point decodings available from Floats, in a slice.
// Any non-parsable items encountered are returned, in the slice, as NaN values, and cause an ErrAnyNaN error to be returned on completion.
func (l *Floats) ReadAll() (fs []float64, err error) {
	fbuf := make([]float64, 100)
	for c := 0; err == nil; {
		c, err = l.Read(fbuf)
		fs = append(fs, fbuf[:c]...)
	}
	if err == io.EOF {
		err = nil
	}
	if err == nil && l.AnyNaN {
		err = ErrAnyNaN{errors.New("Not everything was interpretable as numbers.(returned as NaN)")}
	}
	return
}

// ReadCounter, like Read, reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, an error occurs or the buffer is full.
// But unlike Read it also increments an int, by the number of reads it has made.
// It can be used to find the byte position of a parse failure, this is done by using on a Float with a unit buffer size. This is only intended for testing data sets and/or for retrospective location, due to its lack of buffering giving poor performance.
func (l *Floats) ReadCounter(fs []float64, pos *int) (c int, err error) {
	for c == 0 && err == nil {
		c, err = l.Read(fs)
		*pos = *pos + 1
	}
	return
}

// Read reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, the buffer is full or an error occurs.
func (l *Floats) Read(fs []float64) (c int, err error) {
	var power10 func(uint64) float64
	power10 = func(n uint64) float64 {
		switch n {
		case 0:
			return 1
		case 1:
			return 1e1
		case 2:
			return 1e2
		case 3:
			return 1e3
		case 4:
			return 1e4
		case 5:
			return 1e5
		case 6:
			return 1e6
		case 7:
			return 1e7
		case 8:
			return 1e8
		case 9:
			return 1e9
		default:
			return 1e10 * power10(n-10)
		}
	}
	// function that assembles parsed value and puts it into target float slice
	var setVal func()
	setVal = func() {
		switch l.stage {
		case nan, exponentSign:
			l.AnyNaN = true
			fs[c] = math.NaN()
		case inWhole, beginFraction:
			fs[c] = float64(l.whole)
		case inFraction:
			fs[c] = float64(l.whole) + float64(l.fraction)/power10(uint64(l.fractionDigits))
		default:
			if l.negExponent {
				fs[c] = (float64(l.whole) + float64(l.fraction)/power10(uint64(l.fractionDigits))) / power10(l.exponent)
			} else {
				fs[c] = (float64(l.whole) + float64(l.fraction)/power10(uint64(l.fractionDigits))) * power10(l.exponent)
			}
		}
		if l.neg {
			fs[c] = -fs[c]
		}
		l.whole = 0
		l.fraction = 0
		l.fractionDigits = 0
		l.exponent = 0
		l.neg = false
		l.negExponent = false
		c++
	}
	var n int
	var b []byte
	if len(l.UnBuf) > 0 { // use any unprocessed first, but skip first byte, its just saved separator
		n = len(l.UnBuf)
		b = l.UnBuf
		l.UnBuf = l.UnBuf[0:0]
	} else {
		n, err = l.Reader.Read(l.buf)
		b = l.buf
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			switch l.stage {
			case begin,start:
				l.stage = inWhole
				l.whole = uint64(b[i]) - 48
			case inWhole:
				if l.whole > maxUint {
					l.stage = nan
				} else {
					l.whole *= 10
					l.whole += uint64(b[i]) - 48
				}
			case beginFraction:
				l.stage = inFraction
				l.fraction = uint64(b[i]) - 48
				l.fractionDigits = 1
			case inFraction:
				if l.fraction > maxUint {
					l.stage = nan
				} else {
					l.fraction *= 10
					l.fraction += uint64(b[i]) - 48
					l.fractionDigits++
				}
			case exponentSign:
				l.stage = inExponent
				fallthrough
			case inExponent:
				if l.exponent > maxUint {
					l.stage = nan
				} else {
					l.exponent *= 10
					l.exponent += uint64(b[i]) - 48
				}
			}
		case '.':
			switch l.stage {
			case begin,start, inWhole:
				l.stage = beginFraction
			default:
				l.stage = nan
			}
		case 'e', 'E':
			switch l.stage {
			case inWhole, inFraction:
				l.stage = exponentSign
			default:
				l.stage = nan
			}
		case l.Delimiter: // single delimiter
			//fmt.Println(l)
			switch l.stage {
			case begin:
				l.stage = nan
			case start:
				l.stage = begin
			case exponentSign:
				l.stage = nan
				fallthrough
			default:
				setVal()
				l.stage=begin
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			}
		case ' ', '\n', '\r', '\t', '\f': // delimiters but multiple occurrences only count as one.
			switch l.stage {
			case exponentSign:
				l.stage = nan
				fallthrough
			case inWhole, inFraction, inExponent, beginFraction:
				setVal()
				l.stage=start
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			}
		case '-':
			switch l.stage {
			case begin,start:
				l.neg = true
				l.stage = inWhole
			case exponentSign:
				l.negExponent = true
				l.stage = inExponent
			default:
				l.stage = nan
			}
		case '+':
			switch l.stage {
			case begin,start:
				l.stage = inWhole
			case exponentSign:
				l.stage = inExponent
			default:
				l.stage = nan
			}

		default:
			l.stage = nan
		}
	}
	if err != nil && l.stage != begin && l.stage != start{
		setVal()
		l.stage=begin
	}
	return c, err
}

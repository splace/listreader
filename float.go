package listreader

import "io"
import "math"
import "bytes"
import "errors"

//  TODO could we have a scaled int, so fixed precision rather than floating, version of this?

type progress uint8

const (
	begin progress = iota
	inMultiDelim
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
	fractionDigits uint8  // count of fractional section digits, used to turn interger, into required real, by power of ten division
	exponent       uint64 // exponent section so far read
	negExponent    bool
	AnyNaN         bool   // set if any parsing issue
	buf            []byte // internal buffer.
	UnBuf          []byte // slice of buf of the unconsumed bytes after last Read.
}

// NewFloats returns a Floats reading items from r separated by d, with the bytes package default buffer size.
func NewFloats(r io.Reader, d byte) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bytes.MinRead)}
}

// NewFloatsSize returns a Floats reading items from r separated by d, with an internal buffer size of bSize.
func NewFloatsSize(r io.Reader, d byte, bSize int) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bSize)}
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
		err = ErrAnyNaN
	}
	return
}

var ErrAnyNaN = errors.New("Not everything was interpretable as numbers.(returned as NaN)")

// ReadCounter, like Read, reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, an error occurs or the buffer is full.
// But unlike Read it also returns the number of underlying bytes read.
// It can be used to find the byte position of a parse failure, this is done by using on a Float with a buffer size of one. This is only intended for testing data sets and/or for retrospective location, due to its lack of buffering giving it poor performance.
func (l *Floats) ReadCounter(fs []float64) (c, pos int, err error) {
	for c == 0 && err == nil {
		c, err = l.Read(fs)
		pos++
	}
	return
}

// Read reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, the buffer is full or an error occurs.
// internal buffering means the underlying io.Reader will in general be read past the location of the returned values. (unless the internal buffer length is set to 1.)
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
			case begin, inMultiDelim:
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
			case begin, inMultiDelim, inWhole:
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
			case inMultiDelim:
				l.stage = begin
			case exponentSign:
				l.stage = nan
				fallthrough
			default:
				setVal()
				l.stage = begin
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
				l.stage = inMultiDelim
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			}
		case '-':
			switch l.stage {
			case begin, inMultiDelim:
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
			case begin, inMultiDelim:
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
	if err != nil && l.stage != begin && l.stage != inMultiDelim {
		setVal()
		l.stage = begin
	}
	return c, err
}


// SequenceReaders Read from the embedded Reader until a delimiter, at which point they return with io.EOF.
// to enable Reading on to the next delimiter call Next()
// when reached the io.EOF of the embedded Reader they report EOA (End of All.)
type SequenceReader struct{
	r io.Reader
	delimiter byte
	SectionEnded bool
} 

// Reader compliant Read method. 
func (d *SequenceReader) Read(p []byte) (n int, err error){
	if d.SectionEnded {return 0,io.EOF}
	var c int
	for n=range(p){
		c,err=d.r.Read(p[n:n+1])
		if c==1 && p[n]==d.delimiter{
			d.SectionEnded=true
			return n-1, io.EOF
		}
		if err!=nil{
			break
		}
	}
	if err==io.EOF{
		err=EOA
	}
	return
}

func (d *SequenceReader) Next(){
	d.SectionEnded=false
}
var EOA = errors.New("No More Sections")


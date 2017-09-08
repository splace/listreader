package listreader

import "io"
import "math"
import "bytes"
import "strconv"

type progress uint8

const (
	start progress = iota
	delimFound
	inWhole
	startFraction
	inFraction
	exponentSign
	inExponent
	errorDot
	errorExp
	errorNothing
	errorSign
	errorNondigit
	errorPrecisionLimited
	errorFractionPrecisionLimited
	errorExpPrecisionLimited
)

const maxUint64 = math.MaxUint64 / 10
const maxUint16 = math.MaxUint16 / 10

// Floats is the state, and behaviour, of a floating-point Reader.
type Floats struct {
	io.Reader
	Delimiter      byte
	stage          progress // progress stage or error hit.
	neg            bool     // if negative number
	whole          uint64   // whole number section read so far.
	wholeExponent  uint16   // extra zero digits.
	fraction       uint64   // fraction section read so far.
	fractionDigits uint16   // count of fractional section digits.
	exponent       uint16   // exponent section so far read.
	negExponent    bool     // if exponent negative
	buf            []byte   // internal buffer.
	UnBuf          []byte   // unconsumed bytes after last Read.
}

// NewFloats returns a Floats reading items from r, with Delimiter set to by d. Buffer size set to the bytes package default buffer size.
func NewFloats(r io.Reader, d byte) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bytes.MinRead)}
}

// NewFloatsSize returns a Floats reading items from r, with Delimiter set to by d, and with a particular internal buffer size.
func NewFloatsSize(r io.Reader, d byte, bSize int) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bSize)}
}

// ReadAll returns all the floating-point decodes available from Floats, in a slice, and the first encountered parse, or any embedded Reader, Error.
func (l *Floats) ReadAll() (fs []float64, err error) {
	fbuf := make([]float64,bytes.MinRead)
	for {
		c, rerr := l.Read(fbuf)
		fs = append(fs, fbuf[:c]...)
		if rerr != nil {
			// if its a parse error only keep the first and keep going
			if _, is := rerr.(ParseError); is {
				if err == nil {
					err = rerr
				}
				continue
			}
			// if its not EOF error then return it
			if rerr != io.EOF {
				err = rerr
			}
			return
		}
	}
	return
}

// ParseError is an error recording the stage during a parse that it became conclusive that the text was invalid for a float.
type ParseError progress

func (pe ParseError) Error() string {
	switch progress(pe) {
	case errorDot:
		return "Extra Dot"
	case errorExp:
		return "Exponent Failure"
	case errorNothing:
		return "Empty Item"
	case errorSign:
		return "Extra Sign"
	case errorNondigit:
		return "Non Numeric/White-space/Delimiter encountered"
	case exponentSign:
		return "No Exponent found"
	case errorPrecisionLimited:
		return "Value precision lost in convertion to float64"
	case errorFractionPrecisionLimited:
		return "Value precision lost in convertion to float64"
	case errorExpPrecisionLimited:
		return "Value precision lost in convertion to float64"
	default:
		return "Unknown:#" + strconv.Itoa(int(pe))
	}
}

// Read reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, the buffer is full or an error on the Reader occurs.
// It doesn't abort for parsing errors, but will return the first encountered as type ParseError{}.
// Any non-parsable items encountered are returned, in the slice, as NaN values.
// Internal buffering means the underlying io.Reader will in general be read past the location of the returned values. (unless the internal buffer length is set to 1.)
func (l *Floats) Read(fs []float64) (c int, err error) {
	// optimisation: power ten of int
	var power10 func(uint16) float64 
	power10 = func(n uint16) float64 {
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
		case errorDot, errorExp, errorNothing, errorSign, errorNondigit, exponentSign, errorExpPrecisionLimited:
			err = ParseError(l.stage)
			fs[c] = math.NaN()
		case errorPrecisionLimited:
			err = ParseError(l.stage)
			fs[c] = float64(l.whole)*power10(l.wholeExponent)
		case inWhole, startFraction:
			fs[c] = float64(l.whole)
		case errorFractionPrecisionLimited:
			err = ParseError(l.stage)
			fallthrough
		case inFraction:
			fs[c] = float64(l.whole) + float64(l.fraction)/power10(l.fractionDigits)
		default:
			if l.negExponent {
				fs[c] = (float64(l.whole) + float64(l.fraction)/power10(l.fractionDigits)) / power10(l.exponent)
			} else {
				fs[c] = (float64(l.whole) + float64(l.fraction)/power10(l.fractionDigits)) * power10(l.exponent)
			}
		}
		if l.neg {
			fs[c] = -fs[c]
		}
		l.whole = 0
		l.wholeExponent=0
		l.fraction = 0
		l.fractionDigits = 0
		l.exponent = 0
		l.neg = false
		l.negExponent = false
		c++
	}
	var n int
	var b []byte
	if len(l.UnBuf) > 0 { // use any unprocessed first
		n = len(l.UnBuf)
		b = l.UnBuf
		l.UnBuf = nil
	} else {
		// don’t override an existing parse error (embedded Reader error will still be available on subsequent call.)
		if err == nil {
			n, err = l.Reader.Read(l.buf)
		} else {
			n, _ = l.Reader.Read(l.buf)
		}
		b = l.buf
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			switch l.stage {
			case errorPrecisionLimited:
				l.wholeExponent++
			case delimFound, start:
				l.stage = inWhole
				l.whole = uint64(b[i]) - 48
			case inWhole:
				l.whole *= 10
				if l.whole > maxUint64 {
					l.stage = errorPrecisionLimited
				} else {
					l.whole += uint64(b[i]) - 48
				}
			case startFraction:
				l.stage = inFraction
				l.fraction = uint64(b[i]) - 48
				l.fractionDigits = 1
			case inFraction:
				if l.fraction > maxUint64 {
					l.stage = errorFractionPrecisionLimited
				} else {
					l.fraction *= 10
					l.fraction += uint64(b[i]) - 48
					l.fractionDigits++
				}
			case exponentSign:
				l.stage = inExponent
				fallthrough
			case inExponent:
				if l.exponent > maxUint16 {
					l.stage = errorExpPrecisionLimited
				} else {
					l.exponent *= 10
					l.exponent += uint16(b[i]) - 48
				}
			}
		case '.':
			switch l.stage {
			case delimFound, start, inWhole:
				l.stage = startFraction
			case startFraction, inFraction, exponentSign, inExponent:
				l.stage = errorDot
			}
		case 'e', 'E':
			switch l.stage {
			case inWhole, inFraction:
				l.stage = exponentSign
			case delimFound, start, startFraction, exponentSign, inExponent:
				l.stage = errorExp
			}
		case l.Delimiter: // single delimiter
			switch l.stage {
			case start:
				l.stage = delimFound
			case exponentSign:
				l.stage = errorExp
				setVal()
				l.stage = delimFound
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			case delimFound:
				l.stage = errorNothing
				fallthrough
			default:
				setVal()
				l.stage = delimFound
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			}
		case ' ', '\n', '\r', '\t', '\f': // delimiters, but multiple occurrences are ignored.
			switch l.stage {
			case start, delimFound:
			case exponentSign:
				l.stage = errorExp
			case inWhole, inFraction, inExponent, startFraction:
				setVal()
				l.stage = start
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			default:
				l.stage = start
			}
		case '-':
			switch l.stage {
			case delimFound, start:
				l.neg = true
				l.stage = inWhole
			case exponentSign:
				l.negExponent = true
				l.stage = inExponent
			case inWhole, inFraction, startFraction, inExponent:
				l.stage = errorSign
			}
		case '+':
			switch l.stage {
			case delimFound, start:
				l.stage = inWhole
			case exponentSign:
				l.stage = inExponent
			case inWhole, inFraction, startFraction, inExponent:
				l.stage = errorSign
			}

		default:
			switch l.stage {
			case delimFound, start, inWhole, startFraction, inFraction, exponentSign, inExponent:
				l.stage = errorNondigit
			}
		}
	}
	// did we have something before the error
	if err != nil && l.stage != start {
		switch l.stage {
		case delimFound:
			l.stage = errorNothing
		case exponentSign:
			l.stage = errorExp
		}
		setVal()
		l.stage = start
	}
	return c, err
}


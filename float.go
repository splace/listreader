package listreader

import "io"
import "math"
import "bytes"
import "errors"
import "strconv"

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
	errorDot
	errorExp
	errorNothing
	errorSign
	errorNondigit
	errorTooLarge
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
	fractionDigits uint8  // count of fractional section digits, used to turn integer, into required real, by power of ten division
	exponent       uint64 // exponent section so far read
	negExponent    bool
	buf            []byte // internal buffer.
	UnBuf          []byte // slice of buf of the unconsumed bytes after last Read.
}

// NewFloats returns a Floats reading items from r separated by d, with the bytes package default buffer size.
func NewFloats(r io.Reader, d byte) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bytes.MinRead)}
}

// NewFloatsSize returns a Floats reading items from r separated by d, with a set internal buffer size.
func NewFloatsSize(r io.Reader, d byte, bSize int) *Floats {
	return &Floats{Reader: r, Delimiter: d, buf: make([]byte, bSize)}
}


// ReadAll returns all the floating-point decodings available from Floats, in a slice, and any parse Error.
func (l *Floats) ReadAll() (fs []float64, err error) {
	fbuf := make([]float64, 100)
	for {
		c, eerr := l.Read(fbuf)
		fs = append(fs, fbuf[:c]...)
		if eerr!=nil{
			 // if its a parse error only keep the first and keep going
			if _,is:=eerr.(ParseError);is{
				if err==nil{err=eerr}
				continue
			}
			if eerr!=io.EOF{err=eerr}
			return
			}
	}
	return
}

type ParseError progress

func (pe ParseError)Error()string{
	switch progress(pe){
	case errorDot:
		return "Extra Dot"
	case errorExp:
		return "Exponent Failure"
	case errorNothing:
		return "Empty Item"
	case errorSign:
		return "Extra Sign"
	case errorNondigit:
		return "Non Numeric/Whitespace/Delimiter encountered"
	case exponentSign:
		return "No Exponent found"
	case errorTooLarge:
		return "Value too large"
	default:
		return "Unknown:#"+strconv.Itoa(int(pe))
	} 		
}

// Read reads delimited items and places their decoded floating-point values into the supplied buffer, until the embedded reader needs to be read again, the buffer is full or an error on the Reader occurs.
// It doesn't stop for parsing errors, but returns, the first encountered as type ParseError{}
// Any non-parsable items encountered are returned, in the slice, as NaN values.
// Internal buffering means the underlying io.Reader will in general be read past the location of the returned values. (unless the internal buffer length is set to 1.)
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
		case errorDot,errorExp, errorNothing,errorSign,errorNondigit,exponentSign:
			if err==nil{err=ParseError(l.stage)}
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
	if len(l.UnBuf) > 0 { // use any unprocessed first
		n = len(l.UnBuf)
		b = l.UnBuf
		l.UnBuf = l.UnBuf[0:0]
	} else {
		// don’t override a parse error 
		if err==nil{
			n, err = l.Reader.Read(l.buf)
		}else{
			n, _ = l.Reader.Read(l.buf) 
		}
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
					l.stage = errorTooLarge
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
					l.stage = errorTooLarge
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
					l.stage = errorTooLarge
				} else {
					l.exponent *= 10
					l.exponent += uint64(b[i]) - 48
				}
			}
		case '.':
			switch l.stage {
			case begin, inMultiDelim, inWhole:
				l.stage = beginFraction
			case beginFraction,inFraction,exponentSign,inExponent:
				l.stage = errorDot
			}
		case 'e', 'E':
			switch l.stage {
			case inWhole, inFraction:
				l.stage = exponentSign
			case begin,inMultiDelim,beginFraction,exponentSign,inExponent:
				l.stage = errorExp
			}
		case l.Delimiter: // single delimiter
			//fmt.Println(l)
			switch l.stage {
			case begin:
				l.stage = errorNothing
			case inMultiDelim:
				l.stage = begin
			case exponentSign:
				l.stage = errorExp
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
				l.stage = errorExp
				fallthrough
			case inWhole, inFraction, inExponent, beginFraction:
				setVal()
				l.stage = inMultiDelim
				if c >= len(fs) {
					l.UnBuf = b[i+1 : n]
					return c, nil
				}
			default:
				l.stage = inMultiDelim
			}
		case '-':
			switch l.stage {
			case begin, inMultiDelim:
				l.neg = true
				l.stage = inWhole
			case exponentSign:
				l.negExponent = true
				l.stage = inExponent
			case inWhole, inFraction,beginFraction,inExponent:
				l.stage = errorSign
			}
		case '+':
			switch l.stage {
			case begin, inMultiDelim:
				l.stage = inWhole
			case exponentSign:
				l.stage = inExponent
			case inWhole, inFraction, beginFraction,inExponent:
				l.stage = errorSign
			}

		default:
			l.stage = errorNondigit
		}
	}
	// make sure we capture last item
	if err == io.EOF && l.stage != begin && l.stage != inMultiDelim {
		setVal()
		l.stage = begin
	}
	return c, err
}


// SequenceReaders Read from the embedded Reader until a delimiter, at which point they return with io.EOF.
// to enable Reading on to the next delimiter call Next()
// when reaching the io.EOF of the embedded Reader they report EOA (End of All.)
type SequenceReader struct{
	io.Reader
	delimiter byte
	sectionEnded bool
} 

// Reader compliant Read method. 
func (dr SequenceReader) Read(p []byte) (n int, err error){
	if dr.sectionEnded {return 0,io.EOF}
	var c int
	for n=range(p){
		c,err=dr.Reader.Read(p[n:n+1])
		if c==1 && p[n]==dr.delimiter{
			dr.sectionEnded=true
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

func (dr *SequenceReader) Next(){
	dr.sectionEnded=false
}

var EOA = errors.New("No More Sections")

// CountingReaders Read from the embedded Reader keeping a total of the number of bytes read.
type CountingReader struct {
	io.Reader
	Total int64 
}

// Reader compliant Read method. 
func (cr CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.Total += int64(n)
	return
}


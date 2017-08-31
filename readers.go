package listreader

import "io"
import "errors"


// SequenceReaders Read from the embedded Reader until a delimiter, at which point they return with io.EOF. (so can be handed of as Readers.)
// To enable Reading on to the next delimiter call Next(), (Count records how many times)
// When reaching the io.EOF of the embedded Reader it returns an EOA error.
// They can be used to split by a single unconditional higher level byte, like newline, and even hierarchically by changing the Delimiter. 
// Each section can be passed to different functions that require a Reader.
// Example: the first line, of a table, could be handled by a string list scanner.
type SectionReader struct{
	io.Reader
	Delimiter byte
	Count int64
	sectionEnded bool
} 

// Reader compliant Read method. 
func (dr SectionReader) Read(p []byte) (n int, err error){
	if dr.sectionEnded {return 0,io.EOF}
	var c int
	for n=range(p){
		c,err=dr.Reader.Read(p[n:n+1])
		if c==1 && p[n]==dr.Delimiter{
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

func (dr *SectionReader) Next(){
	dr.Count++
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


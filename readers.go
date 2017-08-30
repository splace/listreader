package listreader

import "io"
import "errors"


// SequenceReaders Read from the embedded Reader until a delimiter, at which point they return with io.EOF.
// to enable Reading on to the next delimiter call Next(), (Count records how many times)
// when reaching the io.EOF of the embedded Reader they report EOA (End of All.)
type SequenceReader struct{
	io.Reader
	Delimiter byte
	Count int64
	sectionEnded bool
} 

// Reader compliant Read method. 
func (dr SequenceReader) Read(p []byte) (n int, err error){
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

func (dr *SequenceReader) Next(){
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


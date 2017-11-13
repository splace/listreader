package listreader

import "io"
import "bytes"

// PartReader is a Reader that ends, as far as Reader consumers are concerned, when it encounters a particular Delimiter.
// Calling its Next() method reads over the next occurrence of Delimiter, allowing continued reading, and it increments Count.
type PartReader struct {
	io.Reader
	Count uint
	Delimiter    byte
	delimiterFound bool
	unused []byte
}

// Read places bytes, from the embedded Reader, into the provided array, when the Delimiter is encountered an error of io.EOF is returned.
func (dr *PartReader) Read(p []byte) (n int, err error) {
	if dr.delimiterFound {
		return 0, io.EOF
	}
	var c int
	var b byte
	if len(dr.unused)>0{
		o:= dr.unused
		if len(p)<len(dr.unused) {
			o=dr.unused[:len(p)]
		}
		for n,b = range o {
			if b == dr.Delimiter {
				dr.delimiterFound = true
				copy(p[:n],dr.unused)
				dr.unused=dr.unused[n+1:]
				return n, io.EOF
			}
		}
		n=copy(p,dr.unused)
		dr.unused=dr.unused[n:]
	}else{
		c, err = dr.Reader.Read(p)
		for n,b = range p[:c] {
			if b == dr.Delimiter {
				dr.delimiterFound = true
				dr.unused=p[n+1:c]  
				if err!=nil {
					break
				}
				return n, io.EOF
			}
		}
	}
	return
}

// Next Reads and discards from the embedded reader a Delimiter, and anything up to it.
// Any error encountered is returned. 
func (dr *PartReader) Next() (err error) {
	dr.Count++
	if !dr.delimiterFound {
		// read and discard remains of section.
		// TODO could read directly any unused, reducing copying taking place
		buf := make([]byte, bytes.MinRead)
		for {
			_, err = dr.Read(buf)
			if err != nil {
				break
			}
		}
	}
	dr.delimiterFound = false
	return
}

// CountingReader is a Reader that keeps a running total of the number of bytes it reads.
type CountingReader struct {
	io.Reader
	Total int64
}

// Read returns bytes, and errors, from the embedded Reader.
func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.Total += int64(n)
	return
}

// CancelableReader is a Reader that can be stopped.
type CancelableReader struct {
	io.Reader
	Cancel bool
}

// Read returns bytes, and errors, from the embedded Reader, unless Cancel is True, then it returns an n of 0 and an err of io.EOF.
// if the Embedded Reader is also a Closer then Close() is called when Cancel is True.
func (cr CancelableReader) Read(p []byte) (n int, err error) {
	if cr.Cancel {
		if r,is:=cr.Reader.(io.Closer);is{
			r.Close()	
		}
		return 0,io.EOF

	}
	return cr.Reader.Read(p)
}


// ConcatReader returns a Reader that returns bytes from Reader's it receives through the provided channel.
// ConcatReader will block waiting for the first Reader on the channel. 
// When one of this Reader's received Readers returns EOF, subsequent Reads return bytes from the next received Reader, if none is available Read will block.
// If a received Reader returns a non-nil, non-EOF error, Read will return that error.
// The Reader returns EOF when the channel is closed.
// If a received Reader is a ReadCloser then close is called on it when it returns an EOF.
func ConcatReader(r <-chan io.Reader) io.Reader{
	return &concatReader{readers:r,currentReader: <-r}
}

type concatReader struct{
	readers <-chan io.Reader
	currentReader io.Reader
}

func (cr *concatReader) Read(p []byte) (n int, err error){
	n,err=cr.currentReader.Read(p)
	if err==io.EOF{
		if n>0 {
			err=nil
			return
			}
		if crc,ok:=cr.currentReader.(io.ReadCloser);ok{
			crc.Close()
		}
		if ncr,ok := <- cr.readers;ok {
			err = nil
			cr.currentReader=ncr
		}
	}
	return
}


package listreader

import "io"
import "bytes"

// PartReader is a Reader that ends, as far as Reader consumers are concerned, when it finds a particular Delimiter, but that can be restarted. 
type PartReader struct {
	io.Reader
	Delimiter    byte
	delimiterFound bool
	Count uint
	unused []byte
}

// Read returns bytes, and errors, from the embedded Reader, when the Delimiter is encountered an err of io.EOF is returned.
func (dr *PartReader) Read(p []byte) (n int, err error) {
	if dr.delimiterFound {
		return 0, io.EOF
	}
	var c int
	var b byte
	if len(dr.unused)>0{
		for n,b = range dr.unused {
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

// Next Reads from the embedded reader, if needed, until just after Delimiter is found.
// Any error encountered is returned. 
func (dr *PartReader) Next() (err error) {
	dr.Count++
	if !dr.delimiterFound {
		// read and discard remains of section.
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




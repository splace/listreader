package listreader

import "io"
import "bytes"

// PartReaders return the, potentially buffered, result of calling Read on the embedded Reader, on encountering the Delimiter they return an err of io.EOF.
// A call to Next() Reads from the embedded reader, if needed, until just after a delimiter is found.
// Any error, from the embedded Reader, encountered while running Next() is returned. 
type PartReader struct {
	io.Reader
	Delimiter    byte
	delimiterFound bool
	Count uint
	unused []byte
}

// Reader compliant Read method.
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

// CountingReader's Read from the embedded Reader keeping a running total of the number of bytes read.
type CountingReader struct {
	io.Reader
	Total int64
}

// Reader compliant Read method.
func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.Total += int64(n)
	return
}

// CancelableReader's pass back the result of calling Read() on their embedded Reader, except when their Cancel property is True, then they return an n of 0 and an err of io.EOF.
// If the embedded Reader is also a Closer, calling Read() will automatically call Close() on it. (when Cancel is True.)
type CancelableReader struct {
	io.Reader
	Cancel bool
}

// Reader compliant Read method.
func (cr CancelableReader) Read(p []byte) (n int, err error) {
	if cr.Cancel {
		if r,is:=cr.Reader.(io.Closer);is{
			r.Close()	
		}
		return 0,io.EOF

	}
	return cr.Reader.Read(p)
}




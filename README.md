# listreader

Status: (Should work, much testing, but not extensively used on variety of data sets, let me know if you find a problem.)

3 to 10 times faster than naive std lib implementation.(using csv.reader or bufio.Scanner)

100MB/s for me, YRMV.

(update: go1.8 much faster, +100% (std.lib. +10% faster) so now 5 to 15 times faster.)
(update: go1.9b1 much faster, +30% (std.lib. also much faster +40%) so now 4 to 12 times faster.)

not actually optimised, just uses an inline 'flat' implementation, so should be able to be made faster.

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/listreader?status.svg)](https://godoc.org/github.com/splace/listreader) 

Installation:

     go get github.com/splace/listreader

Example: parse file in 3 float CSV groups.
```go
package main

import "os"
import "fmt"
import "github.com/splace/listreader"

func main(){
   	file, err := os.Open(os.Args[1])
	defer file.Close()
	reader := listreader.NewFloats(file,',')
	itemBuf := make([]float64, 3)
	for err, c,f := error(nil),0, 0; err == nil; {
		c, err = reader.Read(itemBuf[f:])
	     	f+=c
     		if f<3 {continue}
     		fmt.Println(itemBuf)
     		f=0
 	}
}
```

# listreader

Status: (Should work, wide unit testing, but not well used.)

Overview/docs: [![GoDoc](https://godoc.org/github.com/splace/listreader?status.svg)](https://godoc.org/github.com/splace/listreader) 

Installation:

     go get github.com/splace/listreader

Example:
```go
package main

import "os"
import "fmt"
import "github.com/splace/listReader"

func main(){
   	file, err := os.Open(os.Args[1])
	defer file.Close()
 	if err != nil {
		panic(err)
	}
	reader := listReader.NewFloats(file,',')
	itemBuf := make([]float64, 3)
	for err, c,f := error(nil),0, 0; err == nil; {
		c, err = reader.Read(itemBuf[f:])
		f+=c
     	if f<3 {continue}
      	f=0
     	fmt.Println(itemBuf)
	}
}
```

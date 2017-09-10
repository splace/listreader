package listreader_test

import "github.com/splace/listreader"

import "strings"
import "fmt"

func ExampleFloats() {
	source:=strings.NewReader("1,2,3\n4,5,6,7")
	lineReader := &listreader.PartReader{Reader:source, Delimiter:'\n'}
	for {
		floatReader := listreader.NewFloats(lineReader, ',')
		fs,err:=floatReader.ReadAll()
		if err != nil {
				fmt.Println(err)
				break
		}
		fmt.Println(lineReader.Count,fs)
		err = lineReader.Next()
		if err != nil {
				break
		}
	}
	// Output:
	// 0 [1 2 3]
	// 1 [4 5 6 7]
}


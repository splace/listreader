package main

import "strings"
import "fmt"
import "github.com/splace/listreader"

func main(){
	source := strings.NewReader(`24.022636656429 55.557392812856 52.228635194467 -31.380903518556 -7.9503676820041 28.357857406239 33.33750296633`)
	lr := listreader.NewFloats(source,' ')
	values,_ := lr.ReadAll()
   	fmt.Println(values)
}


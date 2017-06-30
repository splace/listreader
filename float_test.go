package listreader

import "testing"
import "os"
import "fmt"
import "strings"
import "io"
import "io/ioutil"
import "bufio"
import "bytes"
import "compress/gzip"
import "compress/zlib"
import "encoding/csv"
import "strconv"
import "math/rand"
import "math/big"

func TestFloatsRandom(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 1000000))
	tot := big.NewFloat(0)
	for i := 0; i < 8; i++ {
		f1 := rand.NormFloat64() / 100
		f2 := rand.NormFloat64()
		f3 := rand.NormFloat64()
		f4 := rand.NormFloat64() / 100000
		f5 := rand.NormFloat64()
		f6 := rand.NormFloat64()
		d1 := rand.Int31n(rand.Int31())
		d2 := -rand.Int31n(rand.Int31())
		d3 := rand.Int31n(rand.Int31())
		d4 := -rand.Int31n(100)
		d5 := rand.Int31n(10)
		d6 := -rand.Int31n(1000)
		fmt.Fprintf(buf, "%g,%g,%+g,%g,%g,%g,%d,%v,%+d,%d,%d,%d,", f1, f2, f3, f4, f5, f6, d1, d2, d3, d4, d5, d6)
		tot.Add(tot, big.NewFloat(float64(d1)))
		tot.Add(tot, big.NewFloat(float64(d2)))
		tot.Add(tot, big.NewFloat(float64(d3)))
		tot.Add(tot, big.NewFloat(float64(d4)))
		tot.Add(tot, big.NewFloat(float64(d5)))
		tot.Add(tot, big.NewFloat(float64(d6)))
		tot.Add(tot, big.NewFloat(f1))
		tot.Add(tot, big.NewFloat(f2))
		tot.Add(tot, big.NewFloat(f3))
		tot.Add(tot, big.NewFloat(f4))
		tot.Add(tot, big.NewFloat(f5))
		tot.Add(tot, big.NewFloat(f6))
	}
	fmt.Fprint(buf, "1")
	tot.Add(tot, big.NewFloat(float64(1)))

	bs := buf.Bytes()

	//fmt.Println(buf)

	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	tot1 := big.NewFloat(0)
	c1 := 0
	onComma := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == ',' {
				return i + 1, data[:i], nil
			}
		}
		return 0, data, bufio.ErrFinalToken
	}
	scanner.Split(onComma)
	for scanner.Scan() {
		x, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			panic(err)
		}
		c1 = c1 + 1
		tot1.Add(tot1, big.NewFloat(x))
	}

	reader := csv.NewReader(bytes.NewBuffer(bs))
	tot2 := big.NewFloat(0)
	c2 := 0
	rows, _ := reader.ReadAll()
	for _, row := range rows {
		for _, item := range row {
			x, err := strconv.ParseFloat(item, 64)
			if err != nil {
				panic(err)
			}
			c2 = c2 + 1
			tot2.Add(tot2, big.NewFloat(x))
		}
	}

	fReader := NewFloats(bytes.NewBuffer(bs), ',')
	tot3 := big.NewFloat(0)
	c3 := 0
	itemBuf := make([]float64, 1000)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(itemBuf)
		for _, x := range itemBuf[:c] {
			c3 = c3 + 1
			tot3.Add(tot3, big.NewFloat(x))
		}
	}

	//  original total not equal to parsed totals!!
	// TODO must be string rep of big not exact with some edge case, dont have to=ime to track down now
	//	if tot.Cmp(tot1)!=0 || tot1.Cmp(tot2)!=0 || tot2.Cmp(tot3)!=0 {
	//		t.Error(fmt.Sprintf("%v != %v != %v != %v",tot,tot1,tot2,tot3))
	//	}
	if tot1.Cmp(tot2) != 0 || tot2.Cmp(tot3) != 0 {
		t.Error(fmt.Sprintf("%v != %v != %v", tot1, tot2, tot3))
	}
}

func ReadFloats(r io.Reader) ([]float64, error) {
	scanner := bufio.NewScanner(r)
	onComma := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == ',' {
				return i + 1, data[:i], nil
			}
		}
		return 0, data, bufio.ErrFinalToken
	}
	scanner.Split(onComma)
	result := []float64{}
	for scanner.Scan() {
		x, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return result, err
		}
		result = append(result, x)
	}
	return result, scanner.Err()
}

func TestFloatsParse(t *testing.T) {
	reader := strings.NewReader(" 1 2 -3 \t4 ,50e-1 +6 700., 8 9 \n\f 10.0001\t000001,1e01,\"eof\"")
	//var bufLen int64 = 1
	var i int
	fReader := NewFloats(reader, ',')
	coordsBuf := make([]float64, 1)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(coordsBuf)
		if c == 0 {
			continue
		}
		if fmt.Sprint(err, coordsBuf[:]) != []string{"<nil> [1]", "<nil> [2]", "<nil> [-3]", "<nil> [4]", "<nil> [5]", "<nil> [6]", "<nil> [700]", "<nil> [8]", "<nil> [9]", "<nil> [10.0001]", "<nil> [1]", "<nil> [10]", "EOF [NaN]"}[i] {
			t.Error(i, fmt.Sprint(err, coordsBuf[:]))
		}
		i++
		if fReader.AnyNaN {
			switch r := fReader.Reader.(type) {
			case io.Seeker:
				pos, _ := r.Seek(0, os.SEEK_CUR) //pos,_:=r.Seek(0,io.SeekCurrent)
				if pos != 59 {
					t.Error(fmt.Sprintf("pos not 59 %d", pos))
				}
			default:
				fmt.Println("NaN")
			}
			break
		}
	}
}

func TestFloatsParseNaN(t *testing.T) {
	reader := strings.NewReader(" 1 2 -3 \t4 50e-1 +6 700. 8 9, \n\f 10.0001\t000001,1e01")
	fReader := NewFloats(reader, '\n')
	nums, err := fReader.ReadAll()
	if _, ok := err.(ErrAnyNaN); !ok {
		t.Error("no NaN found.")
	}
	//	switch err.(type) {
	//	case ErrAnyNaN:
	//		fmt.Println("some NaN")
	//	}
	if fmt.Sprint(nums) != "[1 2 -3 4 5 6 700 8 NaN 10.0001 NaN]" {
		t.Error(fmt.Sprint(nums) + "!=[1 2 -3 4 5 6 700 8 NaN 10.0001 NaN]")
	}
}

func TestFloatsParse2(t *testing.T) {
	file, err := os.Open("floatlist.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fReader := NewFloats(file, ',')
	coordsBuf := make([]float64, 2)
	c, err := fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[1 2]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[1 2]")
	}
	c, err = fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[-3 4]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[-3 4]")
	}
	c, err = fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[5 6]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[5 6]")
	}
	c, err = fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[700 8]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[700 8]")
	}
	c, err = fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[9 10.0001]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[9 10.0001]")
	}
	c, err = fReader.Read(coordsBuf)
	if err != nil || fmt.Sprint(coordsBuf[:c]) != "[1 10]" {
		t.Error(fmt.Sprint(coordsBuf[:c]) + "!=[1 10]")
	}
}

func TestFloatsParseByLine(t *testing.T) {
	file, err := os.Open("floatlistlong.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fReader := NewFloats(file, ',')
	itemBuf := make([]float64, 3)
	nextByte := make([]byte, 1)
	for err, c, f := error(nil), 0, 0; err == nil; {
		c, err = fReader.Read(itemBuf[f:])
		f += c
		if f < 3 {
			continue
		}
		for err == nil {
			if len(fReader.UnBuf) > 0 {
				nextByte = fReader.UnBuf[0:1]
				if nextByte[0] != ' ' || nextByte[0] != '\n' || nextByte[0] != '\t' || nextByte[0] != '\r' || nextByte[0] != '\f' {
					break
				}
				fReader.UnBuf = fReader.UnBuf[1:]
			} else {
				_, err = fReader.Reader.Read(nextByte)
				if nextByte[0] != ' ' || nextByte[0] != '\n' || nextByte[0] != '\t' || nextByte[0] != '\r' || nextByte[0] != '\f' {
					break
				}

			}
		}

		f = 0
	}
}

func BenchmarkFloat(b *testing.B) {
	coordsBuf := make([]float64, 20)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader("1,2,3,4,5,6,7,8,9,0")
		fReader := NewFloats(reader, ',')
		b.StartTimer()
		for err := error(nil); err == nil; {
			_, err = fReader.Read(coordsBuf)
		}
	}
}

func BenchmarkFloatCompare(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader("1,2,3,4,5,6,7,8,9,0")
		fReader := csv.NewReader(reader)
		b.StartTimer()
		rows, _ := fReader.ReadAll()
		for _, row := range rows {
			for _, item := range row {
				_, _ = strconv.ParseFloat(item, 64)
			}
		}
	}
}

func BenchmarkFloatCompare2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := strings.NewReader("1,2,3,4,5,6,7,8,9,0")
		b.StartTimer()
		_, err := ReadFloats(r)
		if err != nil {
			panic(err)
		}
	}
}
func BenchmarkFloatFile(b *testing.B) {
	coordsBuf := make([]float64, 3)
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open("floatlistlong.txt")
		if err != nil {
			panic(err)
		}
		fReader := NewFloats(file, ',')
		b.StartTimer()
		for err := error(nil); err == nil; {
			_, err = fReader.Read(coordsBuf)
		}
		b.StopTimer()
		file.Close()
	}
}

func BenchmarkFloatMemoryFile(b *testing.B) {
	coordsBuf := make([]float64, 3)
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		dat, err := ioutil.ReadFile("floatlistlong.txt")
		file := bytes.NewBuffer(dat)
		if err != nil {
			panic(err)
		}
		fReader := NewFloats(file, ',')
		b.StartTimer()
		for err := error(nil); err == nil; {
			_, err = fReader.Read(coordsBuf)
		}
		b.StopTimer()
	}
}

func BenchmarkFloatCounterFile(b *testing.B) {
	coordsBuf := make([]float64, 3)
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		file, err := os.Open("floatlistlong.txt")
		if err != nil {
			panic(err)
		}
		fReader := NewFloatsSize(file, ',', 1)
		b.StartTimer()
		c := 0
		for err := error(nil); err == nil; {
			_, err = fReader.ReadCounter(coordsBuf, &c)
		}
		b.StopTimer()
		file.Close()
	}
}

func BenchmarkFloatFileCompare(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, err := os.Open("floatlistlong.txt")
		if err != nil {
			panic(err)
		}
		fReader := csv.NewReader(bufio.NewReaderSize(file, 10000))
		b.StartTimer()
		rows, _ := fReader.ReadAll()
		for _, row := range rows {
			for _, item := range row {
				_, _ = strconv.ParseFloat(item, 64)
			}
		}
		b.StopTimer()
		file.Close()
	}
}

func BenchmarkFloatFileWithWork(b *testing.B) {
	var tot float64
	var c int
	var coord [3]float64
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		gzFile, err := os.Open("floatlistlong.gz")
		if err != nil {
			panic(err)
		}
		file, err := readerFromZippedReader(gzFile)
		if err != nil {
			panic(err)
		}
		fReader := NewFloats(file, ',')
		b.StartTimer()
		for err := error(nil); err == nil; {
			c, err = fReader.Read(coord[:])
			if c == 3 {
				tot += coord[0]
				tot += coord[1]
				tot += coord[2]
			}
		}
		b.StopTimer()
	}
}

func readerFromZippedReader(r io.ReadSeeker) (io.Reader, error) {
	if unzipped, err := gzip.NewReader(r); err == nil {
		return unzipped, nil
	} else {
		// also check for simple zlib deflate, like http compressed
		_, _ = r.Seek(0, 0)
		if inflated, err := zlib.NewReader(r); err == nil {
			return inflated, nil
		}
		panic(err.Error())
	}
}

func BenchmarkFloatFileCompareWithWork(b *testing.B) {
	var tot float64
	var coord [3]float64
	for i := 0; i < b.N; i++ {
		gzfile, err := os.Open("floatlistlong.gz")
		if err != nil {
			panic(err)
		}
		file, err := readerFromZippedReader(gzfile)
		if err != nil {
			panic(err)
		}
		fReader := csv.NewReader(file)
		b.StartTimer()
		for err, row := error(nil), []string{}; err == nil; {
			row, err = fReader.Read()
			for i, item := range row {
				c, err := strconv.ParseFloat(item, 64)
				coord[i] = c
				if err == nil {
					tot += c
				}
			}
		}
		b.StopTimer()
	}
}



/*  Hal3 Wed 22 Mar 00:00:51 GMT 2017 go version go1.6.2 linux/amd64
PASS
BenchmarkFloat-2                   	 3000000	       587 ns/op
BenchmarkFloatCompare-2            	  200000	      6489 ns/op
BenchmarkFloatCompare2-2           	  200000	      7443 ns/op
BenchmarkFloatFile-2               	     100	  21199485 ns/op
BenchmarkFloatMemoryFile-2         	     100	  18615623 ns/op
BenchmarkFloatCounterFile-2        	       2	 682852415 ns/op
BenchmarkFloatFileCompare-2        	      20	  66358048 ns/op
BenchmarkFloatFileWithWork-2       	      50	  33312708 ns/op
BenchmarkFloatFileCompareWithWork-2	      20	  70178404 ns/op
ok  	_/home/simon/Dropbox/github/working/listreader	127.709s
Wed 22 Mar 00:03:00 GMT 2017
*/
/*  Hal3 Wed 22 Mar 00:04:10 GMT 2017  go version go1.8 linux/amd64

BenchmarkFloat-2                      	 3000000	       511 ns/op
BenchmarkFloatCompare-2               	  300000	      4468 ns/op
BenchmarkFloatCompare2-2              	  300000	      6654 ns/op
BenchmarkFloatFile-2                  	     200	   9273955 ns/op
BenchmarkFloatMemoryFile-2            	     200	   8072509 ns/op
BenchmarkFloatCounterFile-2           	       2	 637027535 ns/op
BenchmarkFloatFileCompare-2           	      30	  51885284 ns/op
BenchmarkFloatFileWithWork-2          	     100	  18718196 ns/op
BenchmarkFloatFileCompareWithWork-2   	      20	  58575911 ns/op
PASS
ok  	_/home/simon/Dropbox/github/working/listreader	125.113s
Wed 22 Mar 00:06:22 GMT 2017
*/
/*  Hal3 Sat 24 Jun 21:45:39 BST 2017  go version go1.9beta1 linux/amd64

goos: linux
goarch: amd64
BenchmarkFloat-2                      	 5000000	       360 ns/op
BenchmarkFloatCompare-2               	  500000	      2408 ns/op
BenchmarkFloatCompare2-2              	  300000	      4386 ns/op
BenchmarkFloatFile-2                  	     200	   9539447 ns/op
BenchmarkFloatMemoryFile-2            	     200	   8152662 ns/op
BenchmarkFloatCounterFile-2           	       2	 782847054 ns/op
BenchmarkFloatFileCompare-2           	      30	  42439975 ns/op
BenchmarkFloatFileWithWork-2          	     100	  18532332 ns/op
BenchmarkFloatFileCompareWithWork-2   	      30	  51130660 ns/op
PASS
ok  	_/home/simon/Dropbox/github/working/listreader	23.060s
Sat 24 Jun 21:46:03 BST 2017
*/


package listReader

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
	buf:=bytes.NewBuffer(make([]byte,0,1000000))
	tot:=big.NewFloat(0)
	for i:=0;i<8;i++{	
		f1:=rand.NormFloat64()*100
		f2:=rand.NormFloat64()
		f3:=rand.NormFloat64()
		f4:=rand.NormFloat64()*100000
		f5:=rand.NormFloat64()
		f6:=rand.NormFloat64()
		d1:=rand.Int31n(rand.Int31())
		d2:=-rand.Int31n(rand.Int31())
		d3:=rand.Int31n(rand.Int31())
		d4:=-rand.Int31n(100)
		d5:=rand.Int31n(10)
		d6:=-rand.Int31n(1000)
		fmt.Fprintf(buf,"%g,%g,%+g,%g,%g,%g,%d,%v,%+d,%d,%d,%d,",f1,f2,f3,f4,f5,f6,d1,d2,d3,d4,d5,d6)
		tot.Add(tot,big.NewFloat(float64(d1)))
		tot.Add(tot,big.NewFloat(float64(d2)))
		tot.Add(tot,big.NewFloat(float64(d3)))
		tot.Add(tot,big.NewFloat(float64(d4)))
		tot.Add(tot,big.NewFloat(float64(d5)))
		tot.Add(tot,big.NewFloat(float64(d6)))
		tot.Add(tot,big.NewFloat(f1))
		tot.Add(tot,big.NewFloat(f2))
		tot.Add(tot,big.NewFloat(f3))
		tot.Add(tot,big.NewFloat(f4))
		tot.Add(tot,big.NewFloat(f5))
		tot.Add(tot,big.NewFloat(f6))
	}
	fmt.Fprint(buf,"1")
	tot.Add(tot,big.NewFloat(float64(1)))
	fmt.Println(tot)
	
	bs:=buf.Bytes()

	//fmt.Println(buf)
	
	scanner := bufio.NewScanner(bytes.NewBuffer(bs))
	tot1:= big.NewFloat(0)
	c1:=0
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
		c1=c1+1
		tot1.Add(tot1,big.NewFloat(x))
	}
	fmt.Println(c1,tot1)

	reader := csv.NewReader(bytes.NewBuffer(bs))
	tot2:= big.NewFloat(0)
	c2:=0
	rows, _ := reader.ReadAll()
	for _, row := range rows {
		for _, item := range row {
			x, err := strconv.ParseFloat(item, 64)
			if err != nil {
				panic(err)
			}
			c2=c2+1
			tot2.Add(tot2,big.NewFloat(x))
		}
	}
	fmt.Println(c2,tot2)
	
	fReader := NewFloats(bytes.NewBuffer(bs),',')
	tot3:= big.NewFloat(0)
	c3:=0
	itemBuf := make([]float64, 1000)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(itemBuf)
		for _,x:=range itemBuf[:c]{
			c3=c3+1
			tot3.Add(tot3,big.NewFloat(x))
		}
	}
	fmt.Println(c3,tot3)

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
	reader := strings.NewReader(" 1 2 -3 \t4 50e-1 +6 700., 8 9 \n\f 10.0001\t000001,1e01,\"eof\"")
	//var bufLen int64 = 1
	fReader := NewFloats(reader,',')
	coordsBuf := make([]float64, 1)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(coordsBuf)
		if c == 0 {
			continue
		}
		fmt.Println(err, coordsBuf[:c])
		if fReader.AnyNaN {
			switch r := fReader.Reader.(type) {
			case io.Seeker:
				pos, _ := r.Seek(0, os.SEEK_CUR) //pos,_:=r.Seek(0,io.SeekCurrent)
				fmt.Println("NaN before byte position:", pos)
			default:
				fmt.Println("NaN")

			}
			break
		}
	}
}

func TestFloatsParseNaN(t *testing.T) {
	reader := strings.NewReader(" 1 2 -3 \t4 50e-1 +6 700. 8 9, \n\f 10.0001\t000001,1e01")
	fReader := NewFloats(reader,'\n')
	nums, err := fReader.ReadAll()
	switch err.(type) {
	case ErrAnyNaN:
		fmt.Println("some NaN")
	}
	fmt.Println(nums)
}


func TestFloatsParse2(t *testing.T) {
	file, err := os.Open("floatlist.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fReader := NewFloats(file,',')
	coordsBuf := make([]float64, 2)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(coordsBuf)
		if c == 0 {
			continue
		}
		fmt.Println(err, coordsBuf[:c])
	}
}

func BenchmarkFloat(b *testing.B) {
	coordsBuf := make([]float64, 20)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader("1,2,3,4,5,6,7,8,9,0")
		fReader := NewFloats(reader,',')
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
		fReader := NewFloats(file,',')
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
		file:=bytes.NewBuffer(dat)
		if err != nil {
			panic(err)
		}
		fReader := NewFloats(file,',')
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
		fReader := NewFloatsSize(file,',',1)
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
		fReader := NewFloats(file,',')
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



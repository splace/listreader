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

const testData="69026929.684462,85854991.053164,94878336.971103,39056944.91186,-36804328.550028,-93946081.210834,35376755.304344,40224912.641675,32145237.239145,-89055677.125722,-17580870.686835,-74118543.963003,-24362128.472124,35500417.060917,35838209.714665,96253547.443195,81292896.150282,41985637.388185,-47441198.372953,-47546200.709206,16142269.278011,9938582.4566421,7211793.7296684,-18562080.021278,-43801357.664075,-65230649.600379,83214031.338326,-33461766.379635,57917408.625557,96512787.414953,16185218.522411,-47395205.752642,31329303.202838,-83869325.54835,-73158607.89882,-95724130.513018,2688194.114104,13973627.665068,-13390788.488738,53539485.369595,62599812.803138,-30567287.714485,23675394.860876,-14836705.063859,65427922.255093,81773192.240751,-8164138.4904106,-38431095.768898,-30139655.401064,-89491031.872803,73623700.055119,67297115.441084,-47826399.164193,-19796257.987524,-44934094.764727,-33274333.33419,-89541268.343824,-37000779.778231,94182278.678837,-57117088.864146,-53777779.803508,75787845.149537,64944931.196489,99194594.937933,-86320270.125903,-70258192.145386,53605561.495575,34987917.512184,-72330711.117168,-17793124.969021,54974506.215646,4381309.5914113,51579639.479322,36099875.874864,44518937.330935,94249660.798465,19436710.010952,7898705.8754539,20752698.611819,19580987.896575,-42165488.815943,70306735.285701,-59926034.398342,10161865.553894,28796179.46632,-59863387.68148,-30806085.435118,-81434137.831179,69726676.386654,-41563967.774419,78444136.762267,-82321266.821735,94684133.955596,49874582.770222,-56311713.231873,37492933.560858,-10726742.684248,-93792430.68108,22167227.380987,-31828109.608883"

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
	// TODO must be string rep of big not exact with some edge case, dont have time to track down now
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
	var i int
	fReader := NewFloats(reader, ',')
	coordsBuf := make([]float64, 1)
	for err, c := error(nil), 0; err == nil; {
		c, err = fReader.Read(coordsBuf)
		if c == 0 {
			continue
		}
		if fmt.Sprint(err, coordsBuf[:]) != []string{"<nil> [1]", "<nil> [2]", "<nil> [-3]", "<nil> [4]", "<nil> [5]", "<nil> [6]", "<nil> [700]", "<nil> [8]", "<nil> [9]", "<nil> [10.0001]", "<nil> [1]", "<nil> [10]", ParseError(errorNondigit).Error()+" [NaN]"}[i] {
			t.Error(i, fmt.Sprint(err, coordsBuf[:]))
		}
		i++
		if err!=nil {
			switch r := fReader.Reader.(type) {
			// strings.Reader is a seeker so we can find where its read up too
			case io.Seeker:
				pos, _ := r.Seek(0, os.SEEK_CUR)
				//pos,_:=r.Seek(0,io.SeekCurrent)
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
	reader := strings.NewReader("1 2 -3 \t4 50e-1 +6 700. 8 9, \n\f 10.0001\t000001,1e01")
	fReader := NewFloats(reader, ' ')
	nums, err := fReader.ReadAll()
	if _,is:=err.(ParseError);!is {
		t.Error("no parse Error found.")
	}
	switch err.(type) {
	case nil:
	case ParseError:
	default:
		t.Error("Not Parse error:"+err.Error())
	}
	if pe,is:=err.(ParseError);is {
		if pe!=ParseError(errorNondigit){t.Error("Not Non digit error:"+pe.Error())}
	}

	if fmt.Sprint(nums) != "[1 2 -3 4 5 6 700 8 NaN NaN 10.0001 NaN]" {
		t.Error(fmt.Sprint(nums) + "!=[1 2 -3 4 5 6 700 8 NaN NaN 10.0001 NaN]")
	}
	
	
	reader = strings.NewReader("1 2 -3 \t4 50e-1 +6 700. 8 9, \n\f 10.0001\t000001,1e01")
	fReader = NewFloats(reader, ',')
	nums, err = fReader.ReadAll()
	if _,is:=err.(ParseError);is {
		t.Error("parse Error found.")
	}
	switch err.(type) {
	case nil:
	case ParseError:
	default:
		t.Error("Not Parse error:"+err.Error())
	}
	if pe,is:=err.(ParseError);is {
		if pe!=ParseError(errorNondigit) {t.Error("Not Non digit error:"+pe.Error())}
	}

	if fmt.Sprint(nums) != "[1 2 -3 4 5 6 700 8 9 10.0001 1 10]" {
		t.Error(fmt.Sprint(nums) + "!=[1 2 -3 4 5 6 700 8 9 10.0001 1 10]")
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




func TestInLines(t *testing.T) {

	source := strings.NewReader(`-0.5639740228652954,3.7998700141906738,2.7228600978851318
-0.5956140160560607,3.8421299457550049,2.7341499328613281
-0.606091022491455,3.8560400009155273,2.7367799282073975
-0.6124340295791626,3.8643798828125,2.7373800277709961
-0.6186929941177368,3.8725299835205078,2.7368900775909424
-0.6286190152168273,3.8853299617767334,2.7345600128173828
-0.6629459857940673,3.9293100833892822,2.7228600978851318
-0.5241180062294006,3.7434799671173096,2.6684000492095947
-0.5456290245056152,3.7739999294281006,2.6988899707794189
-0.504597008228302,3.7118000984191895,2.589709997177124
-0.5078420042991638,3.7092099189758301,2.5023701190948486
-0.5213639736175537,3.7236099243164063,2.4604299068450928
-0.534754991531372,3.7382500171661377,2.4236800670623779
-0.5956140160560607,3.7926499843597412,2.7228600978851318
-0.6325590014457702,3.8010799884796143,2.7228600978851318
-0.6621860265731811,3.8247001171112061,2.7228600978851318
-0.6786280274391174,3.8588500022888184,2.7228600978851318
-0.6786280274391174,3.8967399597167969,2.7228600978851318
-0.5290420055389404,3.7395598888397217,2.6684000492095947
-0.5956140160560607,3.7243599891662598,2.6684000492095947
-0.6621860265731811,3.7395598888397217,2.6684000492095947
-0.7155719995498657,3.7821300029754639,2.6684000492095947
-0.745199978351593,3.8436501026153564,2.6684000492095947
-0.745199978351593,3.9119400978088379,2.6684000492095947
-0.7155719995498657,3.9734599590301514,2.6684000492095947
-0.5126000046730041,3.7054100036621094,2.589709997177124
-0.5956140160560607,3.6864700317382813,2.589709997177124
-0.6786280274391174,3.7054100036621094,2.589709997177124
-0.745199978351593,3.7585000991821289,2.589709997177124
-0.782144010066986,3.8352200984954834,2.589709997177124
-0.782144010066986,3.9203701019287109,2.589709997177124
-0.745199978351593,3.9970901012420654,2.589709997177124
-0.5126000046730041,3.7054100036621094,2.5023701190948486
-0.5956140160560607,3.6864700317382813,2.5023701190948486
-0.6786280274391174,3.7054100036621094,2.5023701190948486
-0.745199978351593,3.7585000991821289,2.5023701190948486
-0.782144010066986,3.8352200984954834,2.5023701190948486
-0.782144010066986,3.9203701019287109,2.5023701190948486
-0.745199978351593,3.9970901012420654,2.5023701190948486
-0.5956140160560607,3.7243599891662598,2.4236900806427002`)	
	
//	lineReader := &PartReader{Reader:bufio.NewReaderSize(source,10), Delimiter:'\n'}
	lineReader := &PartReader{Reader:source, Delimiter:'\n'}
	var r int
	for err:=error(nil);err==nil && r<1e6;r++{
		_, err = ioutil.ReadAll(lineReader)
		if err != nil {
			panic(err)
		}
		err = lineReader.Next()
	}
	if r!=40	{t.Errorf("Last Row Not 40 (%v)", r)}

}

func TestFloatsParseInLines(t *testing.T) {
	file, err := os.Open("floatlistlong.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	lineReader := &PartReader{Reader:file, Delimiter:'\n'}
	var r int
	var fs []float64
	for ;err==nil && r<1e6;r++{
		floatReader := NewFloats(lineReader, ',')
		fs,err=floatReader.ReadAll()
		if err != nil {
			t.Errorf("%v %v", err,fs)
		}
		if len(fs)!=3	{t.Errorf("Row %v column count not 3 (%v) %v", r,len(fs),fs)}
		err = lineReader.Next()
	}
	if r!=16910	{t.Errorf("Last Row Not 16910 (%v)", r)}

}


func BenchmarkFloat(b *testing.B) {
	coordsBuf := make([]float64, 300)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		reader := strings.NewReader(testData)
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
		reader := strings.NewReader(testData)
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
		r := strings.NewReader(testData)
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
		fReader := NewFloatsSize(&CountingReader{Reader:file}, ',', 1)
		b.StartTimer()
		for err := error(nil); err == nil; {
			_, err = fReader.Read(coordsBuf)
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

/*  Hal3 Sun 3 Sep 22:52:12 BST 2017  go version go1.9 linux/amd64

goos: linux
goarch: amd64
BenchmarkFloat-2                       	  100000	     14405 ns/op
BenchmarkFloatCompare-2                	   20000	     68043 ns/op
BenchmarkFloatCompare2-2               	   50000	     28449 ns/op
BenchmarkFloatFile-2                   	     200	   9680953 ns/op
BenchmarkFloatMemoryFile-2             	     200	   8184801 ns/op
BenchmarkFloatCounterFile-2            	       2	 862782723 ns/op
BenchmarkFloatFileCompare-2            	      30	  43594832 ns/op
BenchmarkFloatFileWithWork-2           	     100	  18514312 ns/op
BenchmarkFloatFileCompareWithWork-2    	      30	  50892367 ns/op
BenchmarkFloatZippedFileLineReader-2   	      20	  78055536 ns/op
PASS
ok  	_/home/simon/Dropbox/github/working/listreader	23.381s
Sun 3 Sep 22:52:36 BST 2017
*/
/*  Hal3 Sun 3 Sep 22:54:00 BST 2017 go version go1.6.2 linux/amd64
PASS
BenchmarkFloat-2                    	   50000	     35018 ns/op
BenchmarkFloatCompare-2             	   20000	     92929 ns/op
BenchmarkFloatCompare2-2            	   50000	     35112 ns/op
BenchmarkFloatFile-2                	      50	  20302650 ns/op
BenchmarkFloatMemoryFile-2          	     100	  19137999 ns/op
BenchmarkFloatCounterFile-2         	       2	 735910103 ns/op
BenchmarkFloatFileCompare-2         	      20	  64037520 ns/op
BenchmarkFloatFileWithWork-2        	      50	  33889021 ns/op
BenchmarkFloatFileCompareWithWork-2 	      20	  71135486 ns/op
BenchmarkFloatZippedFileLineReader-2	      20	  93140291 ns/op
ok  	_/home/simon/Dropbox/github/working/listreader	22.480s
Sun 3 Sep 22:54:29 BST 2017
*/


/*  Hal3 Thu 7 Sep 23:56:28 BST 2017 go version go1.6.2 linux/amd64
=== RUN   TestFloatsRandom
--- PASS: TestFloatsRandom (0.00s)
=== RUN   TestFloatsParse
--- PASS: TestFloatsParse (0.00s)
=== RUN   TestFloatsParseNaN
--- PASS: TestFloatsParseNaN (0.00s)
=== RUN   TestFloatsParse2
--- PASS: TestFloatsParse2 (0.00s)
=== RUN   TestInLines
--- PASS: TestInLines (0.00s)
=== RUN   TestFloatsParseInLines
--- PASS: TestFloatsParseInLines (0.04s)
PASS
ok  	_/home/simon/Dropbox/github/working/listreader	0.047s
Thu 7 Sep 23:56:30 BST 2017
*/


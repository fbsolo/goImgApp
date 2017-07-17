package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
)

	//	Attached is a list of links leading to an image. Read this list of
	//	images and find 3 most prevalent colors in the RGB scheme in each
	//	image. Write the result into a CSV file in a form of
	//
	//		url;color;color;color
	//
	//	Use RGBA and ignore the alpha . . .

//	I used two different hash functions - one for JPEG and
//	one for non-JPEG because I had trouble building a universal
//	hash function . . .

func genHash2(resp io.ReadCloser) (i uint32){

	//	Calculate a hash based on the URL'd file contents.
	//	Don't use encode / decode machinery from an existing
	//	package (JPEG / PNG / etc.) because we don't reliably
	//	know the file format AKA image type. We only know
	//	that we have a file as a starting point - specifically,
	//	the
	//
	//		response.Body
	//
	//	value. We can't assume anything about the file type - only
	//	that we start(ed) with a genuine file.
	//
	//	Transform the response.Body value into a []byte slice
	//	in the localBuf variable, and proceed . . .

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := io.Copy(gz, resp); err != nil {
		fmt.Println(err.Error())
	}
	gz.Close()

	var localBuf = buf.Bytes()
	h := fnv.New32a()
	h.Write(localBuf)
	i = h.Sum32()
	return i
}

func genHash1(imgData image.Image) (i uint32){

	//	This function hashes a genuine JPEG file. Calculate
	//	a hash based on the URL'd file contents. Use this to
	//	build a map to test if the same file shows up again
	//	in the list . . .

	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, imgData, nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	content := buf.Bytes()
	h := fnv.New32a()
	h.Write(content)
	i = h.Sum32()

	return i
}

func genColor(imgData image.Image) (m map[string]uint32){

	//	Calculate the RGB values. See
	//
	//		https://golang.org/pkg/image/#example_

	//	for more . . .

	m = make(map[string]uint32)
	bounds := imgData.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			//	Ignore the alpha "a" value . . .

			r, g, b, _ := imgData.At(x, y).RGBA()

			// A color's RGBA method returns values in the range [0, 65535].
			// Shifting by 12 reduces this to the range [0, 15].

			//	"Map" the color vals into string variable "s" because
			//	we must use "s" three times. Space-delimit the values
			//	to make it easier to identify the component color
			//	values  . . .

			s := (fmt.Sprint(r >> 12) + " " + fmt.Sprint(g >> 12) + " " + fmt.Sprint(b >> 12))

			//	See if the color exists in the map. Increment if found;
			//	insert into the map if not found . . .

			if _, ok := m[s]; ok {
				m[s]++
			} else {
				m[s] = 1
			}
		}
	}

	return m
}

func readCSV() {

	//	Read each line from .\urls.csv one at a time,
	//	run the analysis function on each line, and
	//	then write directly to .\results.csv one line
	//	at a time. This reduces intermediate steps . . .

	readFile, err := os.Open(`.\\urls.csv`)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer readFile.Close()

	r := csv.NewReader(readFile)

	writeFile, err := os.Create(`.\\results.csv`)
	if err != nil {
		fmt.Println(err.Error())
	}

	defer writeFile.Close()

	w := csv.NewWriter(writeFile)
	defer w.Flush()

	//	Initialize the variables once . . .

	var rec []string
	var response *http.Response

	//	The "source" file
	//
	//		urls.csv
	//
	//	has both JPEG (valid) and non-JPEG (invalid)
	//	files. The output file
	//
	//		results.csv
	//
	//	will show each valid file and each invalid
	//	file, each file only one time. To make sure
	//	each file of each type only appears one time,
	//	use separate maps - one map for the valid files
	//	and one map for the invalid files.

	//	Map hm1 will hold the hash values for the
	//	JPG URL's in urls.csv - the map key will
	//	hold the hash value itself and the map
	//	value will hold a boolean. Map hm2 will hold
	//	the hash values for the non-JPG URL's in
	//	urls.csv - the files that the image color
	//	machinery can't analyze. For these maps,
	//	the value will always be "true" . . .

	hm1 := make(map[uint32]bool)
	hm2 := make(map[uint32]bool)

	errStr := make([]string, 1)
	fileStr := make([]string, 1)

	for {
		rec, _ = r.Read()

		if (rec == nil){

			//	"Simulate" a repeat-until when
			//	when the loop parses the last
			//	file in urls.csv. IOW, leave
			//	the for-loop . . .

			break
		}

		//	First, download the image contents. Variable
		//	responseData will hold a []byte slice . . .

		response, err = http.Get(rec[0])

		if err != nil {
			fmt.Println(err.Error())
		}

		//	https://golang.org/src/image/decode_example_test.go -> func Example() . . .

		imgData, err := jpeg.Decode(response.Body)

		if err != nil {

			//	The software found a non-JPEG file. We want to
			//	place the file name and a message about it in
			//	the finished CSV - only one time.

			//	Get a hash value for this file . . .

			i := genHash2(response.Body)
//			fmt.Println("Line 215: rec[0] = ", rec[0])
//			fmt.Println("Line 216: i = ", i)
//			fmt.Println("Line 217 hm2 = ", hm2)

			//	Look for the hash value in map hm2.
			//	If hm2 does not have that hash value,
			//	insert it, and then write a line in
			//	the output file "results.csv" . . .

			if _, ok := hm2[i]; !ok {

				//	Insert into hm2 . . .

				hm2[i] = true
//				fmt.Println("Line 229: rec[0] = ", rec[0])
//				fmt.Println("Line 230: i = ", i)
//				fmt.Println("Line 231: hm2 = ", hm2)

				//	Write to results.csv . . .

				errStr[0] = rec[0] + " has an invalid file format"
				err := w.Write(errStr)
				if err != nil {
					fmt.Println(err.Error())
				}
			}

			//	Grab the next file in urls.csv and proceed . . .

			continue
		}

		//	Close the response body here, to keep it open if
		//	the non-JPEG block needs it. This call "assumes"
		//	no panic happened - IOW the file cleanly opened
		//	and we can cleanly close it . . .

		response.Body.Close()

		//	The software found a JPEG file. Get
		//	a hash value for this file . . .

		i := genHash1(imgData)

		//	Look for the hash value of the JPEG file in hm1.
		//	If not found, the loop found a "new" file that it
		//	has not seen before. In this case, insert that hash
		//	value in hm1 and proceed . . .

		if _, ok := hm1[i]; !ok {

			//	insert into hm1 . . .

			hm1[i] = true

			//	calculate the RGB values . . .

			m := genColor(imgData)

			//	Now, range through map m three times. Each time,
			//	pick out the largest value and remove that value
			//	for the next iteration. Don't bother with that
			//	removal for the last iteration.
			//
			//	With this approach, we won't need to deal with
			//	actually sorting the map. Three passes through
			//	will give us what we want; actual sorting would
			//	use much more resources.
			//
			//	In each record, space-delimit the RGB color
			//	components of each color to make everything
			//	easier to read . . .

			var count uint32

			fileStr[0] = rec[0] + ";"
			color := ""

			//	Encase the ranging loop in
			//	a for-loop that runs three
			//	times . . .

			for i := 0; i < 3; i++ {
				count = 0
				for k, v := range m {
					if v > count {
						color = k
						count = v
					}
				}

				fileStr[0] += color + ";"
//				fileStr[0] += color + " " + fmt.Sprint(count) + ";"

				//	When the loop extracts the third-highest
				//	count color, we won't need to delete it
				//	because we only need those three highest
				//	count colors . . .

				if (i < 2) {
					delete(m, color)
				}
			}

			err = w.Write(fileStr)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}

	//	Close the response body, using defer. At most, one response.Body
	//	should remain open at this point if a panic happened or something
	//	went wrong, so use defer to call the Close() function one time.
	//	This call will happen outside the block that "deals with" a non-JPEG
	//	file . The call additionally happens at the end of the readCSV()
	//	function . . .

	defer response.Body.Close()
}

func main() {
	readCSV()
}
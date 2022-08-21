// App for uploading image using echo4
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/tidwall/gjson"
)

func upload(c echo.Context) error {

	var err error
	var imgFile *os.File
	var metaData *exif.Exif
	var jsonByte []byte
	var jsonString string

	// Read form fields
	name := c.FormValue("name")
	email := c.FormValue("email")

	//-----------
	// Read file
	//-----------

	// Source
	// -Single file, err := c.FormFile("file")
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]

	resStr := "<p><pre>"
	resStr += "Upload report:\n"

	for i, file := range files {
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Destination
		dstFolder := "up/"
		dst, err := os.Create(dstFolder + file.Filename)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return err
		}

		resStr += fmt.Sprintf("\n\nFile %d: %s OK!\n", i, dst.Name())

		// FF -- Parse-EXIF ------------------------ ___--\\

		// imgFile, err = os.Open("sample.jpg")
		imgFile, err = os.Open(dstFolder + file.Filename)
		if err != nil {
			log.Fatal(err.Error())
		}

		metaData, err = exif.Decode(imgFile)
		if err != nil {
			log.Fatal(err.Error())
		}

		jsonByte, err = metaData.MarshalJSON()
		if err != nil {
			log.Fatal(err.Error())
		}

		jsonString = string(jsonByte)
		fmt.Println(jsonString)

		//"DateTimeOriginal":"2022:08:20 15:25:55"

		// fmt.Println("Make: " + gjson.Get(jsonString, "Make").String())
		// fmt.Println("Model: " + gjson.Get(jsonString, "Model").String())
		// fmt.Println("Software: " + gjson.Get(jsonString, "Software").String())
		// fmt.Println("DateTimeOriginal: " + gjson.Get(jsonString, "DateTimeOriginal").String())
		dtStr := gjson.Get(jsonString, "DateTimeOriginal").String()
		fmt.Println("DateTimeOriginal: " + dtStr)
		// => time: 2022:08:20 15:25:55
		resStr += fmt.Sprintf("       => time: %s \n", dtStr)

		// LL __ Parse-EXIF ________________________ ___--//
		//Date Time Map
		dtm := parseDTString(dtStr)

		//<--MAP-->
		// ==> YYYY 2022
		// ==> MM 08
		// ==> DD 20
		// ==> hh 15
		// ==> mm 25
		// ==> ss 55

		fmt.Println("<--MAP-->")
		for k, v := range dtm {
			fmt.Println("==>", k, v)
		}

		YYYY, _ := strconv.Atoi(dtm["YYYY"])
		MM, _ := strconv.Atoi(dtm["MM"])
		DD, _ := strconv.Atoi(dtm["DD"])
		dow := day_of_week(DD, MM, YYYY)
		// <2022-08-20 Sat 17:46>
		resStr += fmt.Sprintf("<%s-%s-%s %s %s:%s>",
			dtm["YYYY"], dtm["MM"], dtm["DD"], dow,
			dtm["hh"], dtm["mm"])

	}
	resStr += "</pre></p>"
	resStr += "<p>"

	resStr += fmt.Sprintf("\n\n 1978-01-05 is %s <br />\n\n", day_of_week(5, 1, 1978))

	resStr += fmt.Sprintf("\n\n 2022-08-20 is %s <br /> \n\n", day_of_week(20, 8, 2022))

	resStr += fmt.Sprintf("\n\n 0000-03-1 is %s <br />\n\n", day_of_week(1, 3, 0000))

	resStr += "</p>"

	return c.HTML(http.StatusOK, fmt.Sprintf(resStr+"<p>Uploaded total  %d files with fields name=%s and email=%s.</p>", len(files), name, email))

}

func parseDTString(dt string) map[string]string {
	dtmap := make(map[string]string)

	fmt.Println("in parseDTString dt:", dt)

	// FF -- OOO ------------------------ ___--\\
	// regexp get YYYY MM DD  hh mm
	// => time: 2022:08:20 15:25:55
	//rex := regexp.MustCompile(`([0-9]{4}):([0-9]{2}):([0-9]{2}) ([0-9]{2}):([0-9]{2}):([0-9]{2})`)
	rex := regexp.MustCompile(`([0-9]{4}):([0-9]{2}):([0-9]{2})\s([0-9]{2}):([0-9]{2}):([0-9]{2})`)
	rslt := rex.FindAllStringSubmatch(dt, -1)
	for i, m := range rslt {
		fmt.Println(i, "--")
		fmt.Printf("Set %d: YYYY: %s\n", i+1, m[1])
		fmt.Printf("          MM: %s\n", m[2])
		fmt.Printf("          DD: %s\n", m[3])
		fmt.Printf("          hh: %s\n", m[4])
		fmt.Printf("          mm: %s\n", m[5])
		fmt.Printf("          ss: %s\n", m[6])

		dtmap["YYYY"] = m[1]
		dtmap["MM"] = m[2]
		dtmap["DD"] = m[3]
		dtmap["hh"] = m[4]
		dtmap["mm"] = m[5]
		dtmap["ss"] = m[6]

	}
	return dtmap
	// LL __ OOO ________________________ ___--//1

}

// modify from RFC3339 Date and Time on the Internet: Timestamps
// https://www.rfc-editor.org/rfc/rfc3339
// The following is a sample C subroutine loosely based on Zeller's
//   Congruence [Zeller] which may be used to obtain the day of the week
//   for dates on or after 0000-03-01:(0000-03-1 is Wednesday)
func day_of_week(day int, month int, year int) string {
	var cent int
	dayofweek := []string{
		"Sun",
		"Mon",
		"Tue",
		"Wed",
		"Thu",
		"Fri",
		"Sat",
	}
	// dayofweek := []string{
	// 	"Sunday",
	// 	"Monday",
	// 	"Tuesday",
	// 	"Wednesday",
	// 	"Thursday",
	// 	"Friday",
	// 	"Saturday",
	// }

	// adjust months so February is the last one
	month -= 2
	if month < 1 {
		month += 12
		year -= 1
	}
	// split by century
	cent = year / 100
	year %= 100

	//(  (26*month-2)/10 +  day + year<365%7=1> +
	//    year/4<leap year> + cent/4   +  5*cent )  %7

	return (dayofweek[((26*month-2)/10+day+year+year/4+cent/4+5*cent)%7])
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	e.Logger.Fatal(e.Start(":1323"))
}

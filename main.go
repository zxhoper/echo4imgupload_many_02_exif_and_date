// App for uploading image using echo4
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

		resStr += fmt.Sprintf("File %d: %s OK!\n", i, dst.Name())

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
		fmt.Println("DateTimeOriginal: " + gjson.Get(jsonString, "DateTimeOriginal").String())
		resStr += fmt.Sprintf("       => time: %s \n", gjson.Get(jsonString, "DateTimeOriginal").String())

		// LL __ Parse-EXIF ________________________ ___--//

	}
	resStr += "</pre></p>"

	return c.HTML(http.StatusOK, fmt.Sprintf(resStr+"<p>Uploaded total  %d files with fields name=%s and email=%s.</p>", len(files), name, email))

}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)

	e.Logger.Fatal(e.Start(":1323"))
}

package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall/js"
)

type Opf struct {
	Metadata Metadata   `xml:"metadata"`
	Manifest []Manifest `xml:"manifest>item"`
	Spine    Spine      `xml:"spine"`
}

type Metadata struct {
	Title       []string     `xml:"title"`
	Language    []string     `xml:"language"`
	Identifier  []Identifier `xml:"identifier"`
	Creator     []Author     `xml:"creator"`
	Subject     []string     `xml:"subject"`
	Description []string     `xml:"description"`
	Publisher   []string     `xml:"publisher"`
	Contributor []Author     `xml:"contributor"`
	Date        []Date       `xml:"date"`
	Type        []string     `xml:"type"`
	Format      []string     `xml:"format"`
	Source      []string     `xml:"source"`
	Relation    []string     `xml:"relation"`
	Coverage    []string     `xml:"coverage"`
	Rights      []string     `xml:"rights"`
	Meta        []Metafield  `xml:"meta"`
}

type Identifier struct {
	Data   string `xml:",chardata"`
	ID     string `xml:"id,attr"`
	Scheme string `xml:"scheme,attr"`
}

type Author struct {
	Data   string `xml:",chardata"`
	FileAs string `xml:"file-as,attr"`
	Role   string `xml:"role,attr"`
}

type Date struct {
	Data  string `xml:",chardata"`
	Event string `xml:"event,attr"`
}

type Metafield struct {
	Name    string `xml:"name,attr"`
	Content string `xml:"content,attr"`
}

type Manifest struct {
	ID           string `xml:"id,attr"`
	Href         string `xml:"href,attr"`
	MediaType    string `xml:"media-type,attr"`
	Fallback     string `xml:"media-fallback,attr"`
	Properties   string `xml:"properties,attr"`
	MediaOverlay string `xml:"media-overlay,attr"`
}

type Spine struct {
	ID              string      `xml:"id,attr"`
	Toc             string      `xml:"toc,attr"`
	PageProgression string      `xml:"page-progression-direction,attr"`
	Items           []SpineItem `xml:"itemref"`
}

type SpineItem struct {
	IDref      string `xml:"idref,attr"`
	Linear     string `xml:"linear,attr"`
	ID         string `xml:"id,attr"`
	Properties string `xml:"properties,attr"`
}

func PrintWASMLoadStatus(this js.Value, args []js.Value) interface{} {
	return "WASM loaded!"
}

func GetEpubInfo(v js.Value, x[]js.Value) any {
	document := js.Global().Get("document")
	epubInput := document.Call("getElementById", "epub_file")

	epubTitle := document.Call("getElementById", "epub_title")
	epubAuthor := document.Call("getElementById", "epub_author")
	epubIdentifier := document.Call("getElementById", "epub_identifier")
	epubPublisher := document.Call("getElementById", "epub_publisher")
	epubDate := document.Call("getElementById", "epub_date")
	epubLanguage := document.Call("getElementById", "epub_language")

	epubInput.Get("files").Call("item", 0).Call("arrayBuffer").Call("then", js.FuncOf(func(v js.Value, x []js.Value) any {
		data := js.Global().Get("Uint8Array").New(x[0])
		dst := make([]byte, data.Get("length").Int())
		fmt.Printf("%v\n", data.Get("length").Int())
		n := js.CopyBytesToGo(dst, data)
		fmt.Printf("%v\n", n)

		zippedFile, err := zip.NewReader(bytes.NewReader(dst), int64(n))
		if err != nil {
			fmt.Fprintf(os.Stderr, "an error occurred: %v", err)
		}

		for _, subFile := range zippedFile.File {
			fileName := subFile.Name
			//fmt.Printf("%s\n", fileName)

			if strings.Contains(fileName, "content.opf") {
				readFile, err := subFile.Open()
				if err != nil {
					fmt.Fprintf(os.Stderr, "an error occurred: %v", err)
				}
				defer readFile.Close()

				buffer, err := io.ReadAll(readFile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "an error occurred: %v", err)
				}
				//fmt.Println(string(buffer))

				var opf Opf
				err = xml.Unmarshal(buffer, &opf)
				if err != nil {
					fmt.Fprintf(os.Stderr, "an error occurred: %v", err)
				}

				title := ""
				if len(opf.Metadata.Title) > 0 {
					title = opf.Metadata.Title[0]
				}

				author := ""
				if len(opf.Metadata.Creator) > 0 {
					author = opf.Metadata.Creator[0].Data
				}

				identifier := ""
				if len(opf.Metadata.Identifier) > 0 {
					identifier = opf.Metadata.Identifier[0].Data
				}

				publisher := ""
				if len(opf.Metadata.Publisher) > 0 {
					publisher = opf.Metadata.Publisher[0]
				}

				date := ""
				if len(opf.Metadata.Date) > 0 {
					date = opf.Metadata.Date[0].Data
				}

				language := ""
				if len(opf.Metadata.Language) > 0 {
					language = opf.Metadata.Language[0]
				}
				
				fmt.Println("--------------------")
				fmt.Printf("%v\n", title)
				fmt.Printf("%v\n", author)
				fmt.Printf("%v\n", identifier)
				fmt.Printf("%v\n", publisher)
				fmt.Printf("%v\n", date)
				fmt.Printf("%v\n", language)
				fmt.Println("--------------------")

				epubTitle.Set("innerText", title)
				epubAuthor.Set("innerText", author)
				epubIdentifier.Set("innerText", identifier)
				epubPublisher.Set("innerText", publisher)
				epubDate.Set("innerText", date)
				epubLanguage.Set("innerText", language)

				break
			}
		}

		return nil
	}))

	return nil
}

func InitializeApp() {
	document := js.Global().Get("document")
	epubInput := document.Call("getElementById", "epub_file")

	epubInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
		GetEpubInfo(v, x);
		return nil
	}))

}

func main() {
	c := make(chan bool)

	js.Global().Set("PrintWASMLoadStatus", js.FuncOf(PrintWASMLoadStatus))
	js.Global().Set("GetEpubInfo", js.FuncOf(GetEpubInfo))

	InitializeApp()

	<-c
}

//
// Compile with:
//    GOOS=js GOARCH=wasm go build -o main.wasm
// or
//    GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm
// last tested on Go 1.18
//

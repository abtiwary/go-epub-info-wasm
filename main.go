package main

import (
	"syscall/js"
	"fmt"
	"bytes"
	"os"
	"io"
	"archive/zip"
	"encoding/xml"
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

func InitializeApp() {
	document := js.Global().Get("document")

	epubInput := document.Call("getElementById", "epub_file")

	epubTitle := document.Call("getElementById", "epub_title")
	epubAuthor := document.Call("getElementById", "epub_author")
	epubIdentifier := document.Call("getElementById", "epub_identifier")
	epubPublisher := document.Call("getElementById", "epub_publisher")
	epubDate := document.Call("getElementById", "epub_date")
	epubLanguage := document.Call("getElementById", "epub_language")

	epubInput.Set("oninput", js.FuncOf(func(v js.Value, x []js.Value) any {
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

				if fileName == "content.opf" {
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
	
					fmt.Println("--------------------")
					fmt.Printf("%v\n", opf.Metadata.Title[0])
					fmt.Printf("%v\n", opf.Metadata.Creator[0].Data)
					fmt.Printf("%v\n", opf.Metadata.Identifier[0].Data)
					fmt.Printf("%v\n", opf.Metadata.Publisher[0])
					fmt.Printf("%v\n", opf.Metadata.Date[0].Data)
					fmt.Printf("%v\n", opf.Metadata.Language[0])
					fmt.Println("--------------------")
					
					epubTitle.Set("innerText", opf.Metadata.Title[0])
					epubAuthor.Set("innerText", opf.Metadata.Creator[0].Data)
					epubIdentifier.Set("innerText", opf.Metadata.Identifier[0].Data)
					epubPublisher.Set("innerText", opf.Metadata.Publisher[0])
					epubDate.Set("innerText", opf.Metadata.Date[0].Data)
					epubLanguage.Set("innerText", opf.Metadata.Language[0])

					break
				}
			}

			return nil
		}))

		return nil
	}))

}

func main() {
	c := make(chan bool)

	js.Global().Set("PrintWASMLoadStatus", js.FuncOf(PrintWASMLoadStatus))
	//js.Global().Set("WASMUpdateBook", js.FuncOf(WASMUpdateBook))

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

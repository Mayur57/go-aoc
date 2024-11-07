package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

func main() {
	day := 18
	year := 2019
	url := fmt.Sprintf("https://adventofcode.com/%v/day/%v", year, day)
	log.Println(url)
	res, err := http.Get(url)

	if err != nil {
		log.Fatalln(err)
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing body: %v\n", err)
        os.Exit(1)
    }

	articleNode := FindArticle(doc)

	fileWriter, err := os.Create("./_temp/article.html")
	if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating the file: %v\n", err)
        os.Exit(1)
    }
    defer fileWriter.Close()
	html.Render(fileWriter, articleNode)

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error writing html temp file: %v\n", err)
        os.Exit(1)
    }

	HTML2MDX()
}

func FindArticle(n *html.Node) *html.Node {
    if n.Type == html.ElementNode && n.Data == "article" {
        return n
    }
    for child := n.FirstChild; child != nil; child = child.NextSibling {
        if found := FindArticle(child); found != nil {
            return found
        }
    }
    return nil
}
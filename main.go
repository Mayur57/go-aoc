package main

import (
    "bytes"
    "flag"
    "fmt"
    "golang.org/x/net/html"
    "io"
    "os"
    "strings"
)

type MDXConverter struct {
    buffer bytes.Buffer
    indent int
}

func NewMDXConverter() *MDXConverter {
    return &MDXConverter{
        buffer: bytes.Buffer{},
        indent: 0,
    }
}

func (c *MDXConverter) convertNode(n *html.Node) {
    if n.Type == html.TextNode {
        text := strings.TrimSpace(n.Data)
        if text != "" {
            c.buffer.WriteString(text)
        }
        return
    }

    if n.Type != html.ElementNode {
        return
    }

    switch n.Data {
    case "article":
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
    
    case "h1", "h2", "h3", "h4", "h5", "h6":
        c.buffer.WriteString("\n\n")
        c.buffer.WriteString(strings.Repeat("#", int(n.Data[1]-'0')))
        c.buffer.WriteString(" ")
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
        c.buffer.WriteString("\n")

    case "p":
        c.buffer.WriteString("\n\n")
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
        c.buffer.WriteString("\n")

    case "strong":
        c.buffer.WriteString("**")
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
        c.buffer.WriteString("**")

    case "em":
        c.buffer.WriteString("*")
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
        c.buffer.WriteString("*")

    case "a":
        c.buffer.WriteString("[")
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            c.convertNode(child)
        }
        c.buffer.WriteString("](")
        for _, attr := range n.Attr {
            if attr.Key == "href" {
                c.buffer.WriteString(attr.Val)
                break
            }
        }
        c.buffer.WriteString(")")

    case "code":
        if parent := n.Parent; parent != nil && parent.Data == "pre" {
            var language string
            for _, attr := range n.Attr {
                if attr.Key == "class" && strings.HasPrefix(attr.Val, "language-") {
                    language = strings.TrimPrefix(attr.Val, "language-")
                    break
                }
            }
            c.buffer.WriteString("\n\n```")
            c.buffer.WriteString(language)
            c.buffer.WriteString("\n")
            for child := n.FirstChild; child != nil; child = child.NextSibling {
                c.buffer.WriteString(child.Data)
            }
            c.buffer.WriteString("\n```\n")
        } else {
            c.buffer.WriteString("`")
            for child := n.FirstChild; child != nil; child = child.NextSibling {
                c.convertNode(child)
            }
            c.buffer.WriteString("`")
        }

    case "ul", "ol":
        c.buffer.WriteString("\n")
        c.indent++
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            if child.Type == html.ElementNode && child.Data == "li" {
                c.buffer.WriteString(strings.Repeat("  ", c.indent-1))
                if n.Data == "ul" {
                    c.buffer.WriteString("- ")
                } else {
                    c.buffer.WriteString("1. ")
                }
                for grandchild := child.FirstChild; grandchild != nil; grandchild = grandchild.NextSibling {
                    c.convertNode(grandchild)
                }
                c.buffer.WriteString("\n")
            }
        }
        c.indent--
    }
}

func (c *MDXConverter) Convert(r io.Reader) (string, error) {
    doc, err := html.Parse(r)
    if err != nil {
        return "", fmt.Errorf("failed to parse HTML: %v", err)
    }

    var article *html.Node
    var findArticle func(*html.Node) *html.Node

    findArticle = func(n *html.Node) *html.Node {
        if n.Type == html.ElementNode && n.Data == "article" {
            return n
        }
        for child := n.FirstChild; child != nil; child = child.NextSibling {
            if found := findArticle(child); found != nil {
                return found
            }
        }
        return nil
    }

    article = findArticle(doc)
    if article == nil {
        return "", fmt.Errorf("no article element found")
    }

    c.convertNode(article)
    return strings.TrimSpace(c.buffer.String()), nil
}

func HTML2MDX() {
    inputFile := flag.String("input", "article.html", "Input HTML file path")
    outputFile := flag.String("output", "output.mdx", "Output MDX file path")
    flag.Parse()

    input, err := os.Open(*inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
        os.Exit(1)
    }
    defer input.Close()

    converter := NewMDXConverter()
    mdx, err := converter.Convert(input)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error converting HTML to MDX: %v\n", err)
        os.Exit(1)
    }

    err = os.WriteFile(*outputFile, []byte(mdx), 0644)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
        os.Exit(1)
    }

    os.Remove(*inputFile)
    fmt.Printf("Successfully converted %s to %s\n", *inputFile, *outputFile)
}
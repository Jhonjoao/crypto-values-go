package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

//Result exported
type Result struct {
	value  string
	coin   string
	symbol string
}

func (r Result) String() string {
	return fmt.Sprint(r.coin, "(", r.symbol, ")", " - ", r.value)
}

func main() {
	urlToProcess := []string{
		"https://coinmarketcap.com/pt-br/currencies/bitcoin/",
		"https://coinmarketcap.com/pt-br/currencies/ethereum/",
		"https://coinmarketcap.com/pt-br/currencies/terra-luna/",
	}

	ini := time.Now()
	r := make(chan Result)
	go scrapListURL(urlToProcess, r)
	for url := range r {
		fmt.Println(url)
	}

	fmt.Println(" \n(Took ", time.Since(ini).Seconds(), "secs)")
}

func scrapListURL(urlToProcess []string, rchan chan Result) {
	defer close(rchan)
	var results = []chan Result{}

	for i, url := range urlToProcess {
		results = append(results, make(chan Result))
		go scrapParallel(url, results[i])
	}

	for i := range results {
		for r1 := range results[i] {
			rchan <- r1
		}
	}
}

func scrapParallel(url string, rchan chan Result) {
	defer close(rchan)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("ERROR: It can't scrap '", url, "'")
	}
	// Close body when function ends
	defer resp.Body.Close()
	body := resp.Body
	htmlParsed, err := html.Parse(body)
	if err != nil {
		fmt.Println("ERROR: It can't parse html '", url, "'")
	}

	var r Result

	div := getFirstElementByClass(htmlParsed, "h2", "sc-1q9q90x-0")
	r.coin = getFirstTextNode(div).Data

	div = getFirstElementByClass(htmlParsed, "small", "nameSymbol")
	r.symbol = getFirstTextNode(div).Data

	div = getFirstElementByClass(htmlParsed, "div", "priceValue")
	r.value = getFirstTextNode(div).Data

	rchan <- r
}

func hasClass(attribs []html.Attribute, className string) bool {
	for _, attr := range attribs {
		if attr.Key == "class" && strings.Contains(attr.Val, className) {
			return true
		}
	}
	return false
}

func getFirstTextNode(htmlParsed *html.Node) *html.Node {
	if htmlParsed == nil {
		return nil
	}

	for m := htmlParsed.FirstChild; m != nil; m = m.NextSibling {
		if m.Type == html.TextNode {
			return m
		}
		r := getFirstTextNode(m)
		if r != nil {
			return r
		}
	}
	return nil
}

func getFirstElementByClass(htmlParsed *html.Node, elm, className string) *html.Node {
	for m := htmlParsed.FirstChild; m != nil; m = m.NextSibling {
		if m.Data == elm && hasClass(m.Attr, className) {
			return m
		}
		r := getFirstElementByClass(m, elm, className)
		if r != nil {
			return r
		}
	}
	return nil
}

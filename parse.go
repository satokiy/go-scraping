package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseList(resp *http.Response) ([]Item, error) {

	body := resp.Body
	requesetURL := resp.Request.URL

	var items []Item

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("error loading http response body: %v", err)
	}

	tr := doc.Find("table tr")
	notFoundMessage := "No results found"
	if strings.Contains(doc.Text(), notFoundMessage) || tr.Size() == 0 {
		return nil, nil
	}

	tr.Each(
		func(_ int, s *goquery.Selection) {
			// item構造体
			item := Item{}
			// name
			item.Name = s.Find("td:nth-of-type(2) a").Text()
			item.Price, _ = strconv.Atoi(
				strings.ReplaceAll(
					strings.ReplaceAll(s.Find("td:nth-of-type(3)").Text(), ",", ""),
					"円", ""))
			itemURL, exists := s.Find("td:nth-of-type(2) a").Attr("href")
			refURL, parseErr := url.Parse(itemURL)

			if exists && parseErr == nil {
				// 絶対URLに変換
				item.URL = (*requesetURL.ResolveReference(refURL)).String()
			}

			if item.Name != "" {
				items = append(items, item)
			}
		})

	return items, nil

}

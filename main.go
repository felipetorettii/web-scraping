package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Junkes887/artifacts"
	"github.com/Junkes887/artifacts/model"
	"github.com/Junkes887/web-scraping/connection"
	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/segmentio/kafka-go"
)

func main() {
	godotenv.Load()
	driver := connection.ConnectionDB()
	defer driver.Close()
	con := connection.ConnectionKafka()
	consume(driver, con)
}

func consume(driver neo4j.Driver, con *kafka.Reader) {
	for {
		menssage, err := con.ReadMessage(context.Background())
		artifacts.HandlerError(err)
		url := string(menssage.Value)

		session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
		defer session.Close()

		res := request(url)
		page := manipulateHTML(res.Body)
		page.Link = url

		created := createNodePage(session, page)
		if created {
			createBindsPage(session, page)
		}
		fmt.Println("Create page - " + page.Title)
	}
}

func request(url string) *http.Response {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	artifacts.HandlerError(err)

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:58.0) Gecko/20100101 Firefox/58.0")
	req.Header.Set("Accept", "text/html")

	res, err := client.Do(req)
	artifacts.HandlerError(err)

	return res
}

func createNodePage(session neo4j.Session, page model.Page) bool {
	res, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		result, err := tx.Run("MERGE (p:Page{link:$link}) ON CREATE SET p.title = $title, p.description = $description, p.link = $link", map[string]interface{}{
			"title":       page.Title,
			"description": page.Description,
			"link":        page.Link,
		})
		if err != nil {
			log.Panicln(err)
			return nil, err
		}
		return result.Consume()
	})
	artifacts.HandlerError(err)
	return res != nil
}

func createBindsPage(session neo4j.Session, page model.Page) {
	re, err := regexp.Compile("[^\\p{L}\\d\\s]+")
	artifacts.HandlerError(err)
	fDescription := re.ReplaceAllString(strings.ToUpper(page.Description), "")
	fTitle := re.ReplaceAllString(strings.ToUpper(page.Title), "")
	arr := append(strings.Fields(fDescription), strings.Fields(fTitle)...)

	for _, word := range arr {
		if irrelevantWord(word) {
			continue
		}
		session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
			result, err := tx.Run(
				"MATCH (p:Page{link:$link})"+
					"MERGE(p)-[v:Vinculo]->(w:Word {text:$word, frequency:$frequency})", map[string]interface{}{
					"link":      page.Link,
					"word":      word,
					"frequency": strings.Count(" "+fDescription+" "+fTitle+" ", " "+word+" "),
				})
			if err != nil {
				log.Panicln(err)
				return nil, err
			}
			return result.Consume()
		})
	}
}

func irrelevantWord(word string) bool {
	switch word {
	case "A", "E", "I", "O", "U", "DA", "DE", "DI", "DO", "DU", "AS", "NA", "NO":
		return true
	}

	return false
}

func manipulateHTML(res io.ReadCloser) model.Page {
	doc, err := goquery.NewDocumentFromReader(res)
	artifacts.HandlerError(err)
	var description, title string

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		op, _ := s.Attr("property")
		con, _ := s.Attr("content")

		switch op {
		case "og:description", "twitter:description":
			description = con
		case "og:title", "twitter:title":
			title = con
		}
	})

	return model.Page{
		Title:       title,
		Description: description,
	}
}

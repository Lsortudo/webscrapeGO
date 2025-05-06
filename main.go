package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

var URLM = "https://mangapark.net/title/10953-en-one-piece"
var baseURL = "https://mangapark.net"

func main() {
	c := colly.NewCollector()

	// Procura pelo capítulo 1146
	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	// 	if strings.Contains(e.Text, "Ch. 1146") {
	// 		chapterURL := "https://comick.io" + e.Attr("href")
	// 		fmt.Println("Encontrado:", chapterURL)
	// 		goToChapter(chapterURL)
	// 	}
	// })
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		//fmt.Println("Link encontrado:", e.Text, "→", e.Attr("href"))

		if strings.Contains(e.Text, "Vol.TBE Ch.1146") {
			chapterURL := baseURL + e.Attr("href")
			fmt.Println("Encontrado:", chapterURL)
			//goToChapter(chapterURL)
		}
	})

	// Visita a página principal do mangá
	c.Visit(URLM)
}

/* TO DO
Entrar no site (ja especificado qual maanga)
Percorrer a lista toda de todos os capitulos
Entrar nesses capitulos
Baixar as imagens
Salvar como PDF (ou talvez mando individualmente, ver qual é melhor)
Tambem enviar os dados como releaseDate e o nome/title do capitulo, pra poder usar no meu site todas essas informacoes
*/

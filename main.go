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

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		fmt.Println("Link encontrado:", e.Text, "→", e.Attr("href"))

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
X Entrar no site (ja especificado qual maanga)
X Percorrer a lista toda de todos os capitulos
X Listar todos os links (possiveis capitulos que posso pegar)
Entrar nesses capitulos
Baixar as imagens
Salvar como PDF (ou talvez mando individualmente, ver qual é melhor)
Tambem enviar os dados como releaseDate e o nome/title do capitulo, pra poder usar no meu site todas essas informacoes
*/

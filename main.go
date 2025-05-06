package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/signintech/gopdf"
)

var (
	baseURL       = "https://mangapark.net"
	URLM          = "https://mangapark.net/title/10953-en-one-piece"
	targetChapter = "Vol.TBE Ch.1146"
)

func main() {
	var chapterURL string
	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		//fmt.Println("Link encontrado:", e.Text, "→", e.Attr("href"))

		if strings.Contains(e.Text, targetChapter) {
			chapterURL = baseURL + e.Attr("href")
			fmt.Println("Encontrado:", chapterURL)
			//goToChapter(chapterURL)
		}
	})

	// Visita a página principal do mangá
	err := c.Visit(URLM)
	if err != nil {
		log.Fatal("Erro ao visitar página principal:", err)
	}
	if chapterURL == "" {
		log.Fatal("Capítulo não encontrado!")
	}

	// Agora visita o link do capítulo e coleta as imagens
	imageLinks := []string{}
	chapterCollector := colly.NewCollector()

	chapterCollector.OnHTML("img[src]", func(e *colly.HTMLElement) {
		src := e.Attr("src")
		if strings.HasPrefix(src, "https://") && (strings.HasSuffix(src, ".jpg") || strings.HasSuffix(src, ".png")) {
			imageLinks = append(imageLinks, src)
		}
	})

	err = chapterCollector.Visit(chapterURL)
	if err != nil {
		log.Fatal("Erro ao visitar o capítulo:", err)
	}

	if len(imageLinks) == 0 {
		log.Fatal("Nenhuma imagem encontrada.")
	}

	fmt.Printf("Total de imagens: %d\n", len(imageLinks))

	// Baixa imagens e cria o PDF
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for i, imgURL := range imageLinks {
		filename := fmt.Sprintf("page_%d.jpg", i)
		err := downloadImage(imgURL, filename)
		if err != nil {
			log.Println("Erro ao baixar imagem:", imgURL, err)
			continue
		}

		pdf.AddPage()
		//err = pdf.Image(filename, 0, 0, 595.28, 841.89) // A4
		err = pdf.Image(filename, 0, 0, &gopdf.Rect{W: 595.28, H: 841.89})
		if err != nil {
			log.Println("Erro ao adicionar imagem ao PDF:", err)
		}
		os.Remove(filename)
	}

	err = pdf.WritePdf("one_piece_ch_1146.pdf")
	if err != nil {
		log.Fatal("Erro ao salvar PDF:", err)
	}

	fmt.Println("PDF gerado com sucesso!")
}

func downloadImage(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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

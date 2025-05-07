package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/signintech/gopdf"
)

var (
	baseURL       = "https://mangapark.net"
	mainURL       = "https://mangapark.net/title/10953-en-one-piece"
	targetChapter = "Vol.TBE Ch.1146"
)

func main() {
	// Chama a funcao que vai fazer o Scrape
	scrape()
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

func scrape() {
	var chapterURL string

	c := colly.NewCollector()
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Text, targetChapter) {
			chapterURL = baseURL + e.Attr("href")
			fmt.Println("Capítulo encontrado:", chapterURL)
		}
	})
	err := c.Visit(mainURL)
	if err != nil {
		log.Fatal("Erro ao visitar página principal:", err)
	}
	if chapterURL == "" {
		log.Fatal("Capítulo não encontrado!")
	}

	// Acessando a pagina do cap pra pegar as imgs, ja que antes nao taava funcionando com 100% colly (talvez pq as imgs carregavam no JS ou algo do tipo, mas com o chromeDP funcionou)
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var imageLinks []string
	// Executa o Chrome
	err = chromedp.Run(ctx,
		chromedp.Navigate(chapterURL),
		chromedp.WaitVisible(`img`, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('img')).map(img => img.src)`, &imageLinks),
	)
	if err != nil {
		log.Fatal("Erro ao extrair imagens com chromedp:", err)
	}

	if len(imageLinks) == 0 {
		fmt.Println("Nenhuma imagem encontrada.")
		return
	}

	fmt.Printf("Foram encontradas %d imagens.\n", len(imageLinks))

	// Temp pasta
	os.Mkdir("images", os.ModePerm)

	var filenames []string
	for i, link := range imageLinks {
		filename := fmt.Sprintf("images/page_%03d.jpg", i+1)
		err := downloadImage(link, filename)
		if err == nil {
			filenames = append(filenames, filename)
		} else {
			log.Println("Erro ao baixar imagem:", link, err)
		}
	}

	// Gera o PDF
	err = imagesToPDF(filenames, "one_piece_chapter_1146.pdf")
	if err != nil {
		log.Fatal("Erro ao gerar PDF:", err)
	}

	fmt.Println("PDF gerado com sucesso!")
}

func imagesToPDF(images []string, output string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	for _, img := range images {
		pdf.AddPage()
		err := pdf.Image(img, 0, 0, &gopdf.Rect{W: 595.28, H: 841.89}) // A4
		if err != nil {
			return err
		}
	}
	return pdf.WritePdf(output)
}

/* TO DO
X Entrar no site (ja especificado qual maanga)
X Percorrer a lista toda de todos os capitulos
X Listar todos os links (possiveis capitulos que posso pegar)
X Entrar nesses capitulos
X Baixar as imagens
X Salvar como PDF (ou talvez mando individualmente, ver qual é melhor)
Tambem enviar os dados como releaseDate e o nome/title do capitulo, pra poder usar no meu site todas essas informacoes
*/

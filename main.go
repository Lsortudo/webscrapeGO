package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	baseURL = "https://mangapark.net"
	mainURL = "https://mangapark.net/title/74491-en-blue-lock"
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
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var chapterHref string
	err := chromedp.Run(ctx,
		chromedp.Navigate(mainURL),
		chromedp.WaitVisible(`div.group.flex.flex-col`, chromedp.ByQuery),
		chromedp.AttributeValue(
			`div.group.flex.flex-col a`,
			"href",
			&chapterHref,
			nil,
			chromedp.ByQuery,
		),
	)
	if err != nil {
		log.Fatal("Erro ao encontrar capítulo:", err)
	}
	if chapterHref == "" {
		log.Fatal("Capítulo não encontrado!")
	}

	fullChapterURL := baseURL + chapterHref
	fmt.Println("Último capítulo encontrado:", fullChapterURL)

	var chapterID string
	// Usa expressão regular para extrair os dígitos do final da URL (usar substring do final igual Aldo ensinou)
	re := regexp.MustCompile(`(\d+)$`)
	match := re.FindStringSubmatch(chapterHref)
	if len(match) > 1 {
		chapterID = match[1]
	} else {
		chapterID = "latest"
	}

	var imageLinks []string
	err = chromedp.Run(ctx,
		chromedp.Navigate(fullChapterURL),
		chromedp.WaitVisible(`div.grid-cols-1 div[data-name="image-item"] img`, chromedp.ByQuery),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('div.grid-cols-1 div[data-name="image-item"] img')).map(img => img.src)`, &imageLinks),
	)
	if err != nil {
		log.Fatal("Erro ao extrair imagens com chromedp:", err)
	}

	if len(imageLinks) == 0 {
		fmt.Println("Nenhuma imagem encontrada.")
		return
	}

	fmt.Printf("Foram encontradas %d imagens.\n", len(imageLinks))
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

	outputFile := fmt.Sprintf("one_piece_ch_%s.cbz", chapterID)
	err = createCBZ(filenames, outputFile)
	if err != nil {
		log.Fatal("Erro ao gerar CBZ:", err)
	}

	fmt.Println("CBZ gerado com sucesso:", outputFile)
}

func createCBZ(files []string, output string) error {
	cbzFile, err := os.Create(output)
	if err != nil {
		return err
	}
	defer cbzFile.Close()

	zipWriter := zip.NewWriter(cbzFile)
	defer zipWriter.Close()

	for _, file := range files {
		err := addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}

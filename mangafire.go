package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/jung-kurt/gofpdf"
)

func main() {
	chapter := MangaChapter{}

	c := colly.NewCollector()

	// Visita a página principal do mangá
	c.OnHTML("li.item", func(e *colly.HTMLElement) {
		title := e.ChildAttr("a", "title")
		name := e.ChildText("a span:nth-of-type(1)")
		date := e.ChildText("a span:nth-of-type(2)")
		href := e.ChildAttr("a", "href")
		chNum := e.Attr("data-number")

		// Filtra o capítulo 1
		if chNum == "1" {
			chapter = MangaChapter{
				URL:         "https://mangafire.to" + href,
				Title:       title,
				Name:        name,
				ReleaseDate: date,
				ChapterNum:  chNum,
			}
		}
	})

	err := c.Visit("https://mangafire.to/manga/dandadann.3r5x9")
	if err != nil {
		log.Fatal(err)
	}

	// Agora acessa o capítulo 1 e pega as imagens
	imageCollector := colly.NewCollector()

	imageCollector.OnHTML("div#readerarea img", func(e *colly.HTMLElement) {
		img := e.Attr("src")
		chapter.Images = append(chapter.Images, img)
	})

	err = imageCollector.Visit(chapter.URL)
	if err != nil {
		log.Fatal(err)
	}

	// Baixar as imagens e montar PDF
	err = createPDF(chapter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PDF gerado com sucesso!")
}

func checkImageURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code inválido: %d", resp.StatusCode)
	}

	// Lê os primeiros 512 bytes para detectar o tipo
	buf := make([]byte, 512)
	_, err = io.ReadFull(resp.Body, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("erro ao ler dados: %v", err)
	}

	contentType := http.DetectContentType(buf)
	log.Printf("DEBUG: URL: %s\nTipo detectado: %s\nTamanho: %d bytes\n",
		url, contentType, resp.ContentLength)

	return contentType, nil
}

func createPDF(ch MangaChapter) error {
	// Primeiro fazemos todas as verificações
	for i, imgURL := range ch.Images {
		contentType, err := checkImageURL(imgURL)
		if err != nil {
			return fmt.Errorf("erro na imagem %d (%s): %v", i+1, imgURL, err)
		}

		if !strings.HasPrefix(contentType, "image/") {
			return fmt.Errorf("URL %s não é uma imagem (tipo: %s)", imgURL, contentType)
		}
		log.Printf("Imagem %d válida: %s (%s)\n", i+1, imgURL, contentType)
	}

	// Só criamos o PDF depois de verificar todas as imagens
	pdf := gofpdf.New("P", "mm", "A4", "")

	for i, imgURL := range ch.Images {
		resp, err := http.Get(imgURL)
		if err != nil {
			return fmt.Errorf("erro ao baixar imagem %s: %v", imgURL, err)
		}
		defer resp.Body.Close()

		imgData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("erro ao ler dados da imagem: %v", err)
		}

		// Detecta tipo pelo conteúdo
		imgType := ""
		switch {
		case bytes.HasPrefix(imgData, []byte{0x89, 0x50, 0x4E, 0x47}):
			imgType = "PNG"
		case bytes.HasPrefix(imgData, []byte{0xFF, 0xD8}):
			imgType = "JPEG"
		case bytes.HasPrefix(imgData, []byte{0x47, 0x49, 0x46, 0x38}):
			imgType = "GIF"
		case bytes.HasPrefix(imgData, []byte{0x52, 0x49, 0x46, 0x46}) &&
			len(imgData) > 8 && bytes.Equal(imgData[8:12], []byte{0x57, 0x45, 0x42, 0x50}):
			imgType = "WEBP"
		case bytes.HasPrefix(imgData, []byte{0x25, 0x50, 0x44, 0x46}):
			return fmt.Errorf("a URL %s é um PDF, não uma imagem", imgURL)
		default:
			return fmt.Errorf("tipo de imagem não suportado ou desconhecido: %s", imgURL)
		}

		pdf.AddPage()
		imgName := fmt.Sprintf("img%d", i)

		// Registra a imagem no PDF
		opt := gofpdf.ImageOptions{ImageType: imgType}
		pdf.RegisterImageOptionsReader(imgName, opt, bytes.NewReader(imgData))

		// Obtém informações da imagem
		info := pdf.GetImageInfo(imgName)
		if info == nil {
			return fmt.Errorf("não foi possível processar a imagem %s", imgURL)
		}

		// Calcula dimensões para caber na página
		pageW, pageH := pdf.GetPageSize()
		margin := 10.0 // 10mm de margem
		contentW := pageW - 2*margin
		contentH := pageH - 2*margin

		imgW := info.Width()
		imgH := info.Height()
		ratio := imgW / imgH

		// Redimensiona mantendo proporção
		if imgW > contentW {
			imgW = contentW
			imgH = imgW / ratio
		}
		if imgH > contentH {
			imgH = contentH
			imgW = imgH * ratio
		}

		// Centraliza na página
		x := (pageW - imgW) / 2
		y := (pageH - imgH) / 2

		// Adiciona a imagem ao PDF
		pdf.Image(imgName, x, y, imgW, imgH, false, "", 0, "")
	}

	fileName := sanitizeFilename(ch.Name) + ".pdf"
	err := pdf.OutputFileAndClose(fileName)
	if err != nil {
		return fmt.Errorf("erro ao salvar PDF: %v", err)
	}

	// Verifica se o PDF foi criado
	if fileInfo, err := os.Stat(fileName); err != nil || fileInfo.Size() == 0 {
		return fmt.Errorf("o PDF gerado está vazio ou não foi criado")
	}

	log.Printf("PDF gerado com sucesso: %s (%d páginas)\n", fileName, len(ch.Images))
	return nil
}

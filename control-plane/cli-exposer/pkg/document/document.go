package document

import (
	"fmt"
	"github.com/russross/blackfriday/v2"
	"os"
	"path/filepath"
)

var (
	staticFilePath = "/app/static/"
	cssFilename    = "pico.min.css"
	jsFilename     = "script.js"
)

type HtmlContent struct {
	Js  []byte
	Css []byte
}

func SetStaticFilePath(path string) {
	staticFilePath = path
}

func GetHtmlContent() (HtmlContent, error) {
	var err error
	var cssContent, jsContent []byte

	cssContent, err = os.ReadFile(filepath.Join(staticFilePath, cssFilename))
	if err != nil {
		return HtmlContent{}, fmt.Errorf("error reading css file: %w", err)
	}
	jsContent, err = os.ReadFile(filepath.Join(staticFilePath, jsFilename))
	if err != nil {
		return HtmlContent{}, fmt.Errorf("error reading js file: %w", err)
	}

	return HtmlContent{
		Js:  jsContent,
		Css: cssContent,
	}, nil
}

// GenerateDocPage generates a HTML page from a markdown document
func GenerateDocPage(markdownDoc, cssContent, jsContent []byte) string {
	htmlContent := blackfriday.Run(markdownDoc)
	htmlTemplate := fmt.Sprintf(`
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<meta name="color-scheme" content="light dark" />
			<style>%s</style>
            <title>Researcher Command Line Interface Documentation</title>
        </head>
        <body>
		<main class="container">
		<nav>
		  <ul>
			<li><strong>Researcher Command Line Interface</strong></li>
		  </ul>
		  <ul>
			<li><a href="runai.md">Home</a></li>
			<li><button class="outline secondary" id="changeDarkModeButton">Dark/Light</button></li>
		  </ul>
		</nav>
            %s
		</main>
		<script>%s</script>
        </body>
        </html>
    `, cssContent, htmlContent, jsContent)
	return htmlTemplate
}

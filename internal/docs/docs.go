package docs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/api/docs/v1"
)

// ConvertDocsToMarkdown retrieves a Google Doc and converts its content to markdown format.
func ConvertDocsToMarkdown(docsSrv *docs.Service, docID string, baseFolder string) (string, string, error) {
	fmt.Printf("Retrieving document with ID: %s\n", docID)

	doc, err := docsSrv.Documents.Get(docID).Do()
	if err != nil {
		return "", "", fmt.Errorf("unable to retrieve document: %v", err)
	}

	docFolderPath := filepath.Join(baseFolder, doc.Title)
	err = os.MkdirAll(docFolderPath, 0700)
	if err != nil {
		return "", "", fmt.Errorf("unable to create document folder: %v", err)
	}

	var sb strings.Builder

	// Add the document title as the largest heading
	sb.WriteString("# " + doc.Title + "\n\n")

	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			// Handle paragraph styles (headings, lists)
			if elem.Paragraph.ParagraphStyle != nil {
				if strings.HasPrefix(elem.Paragraph.ParagraphStyle.NamedStyleType, "HEADING_") {
					headingLevel := strings.TrimPrefix(elem.Paragraph.ParagraphStyle.NamedStyleType, "HEADING_")
					headingPrefix := strings.Repeat("#", headingLevelToInt(headingLevel)) + " "
					sb.WriteString(headingPrefix)
				} else if elem.Paragraph.Bullet != nil {
					bulletType := elem.Paragraph.Bullet.ListId
					list := doc.Lists[bulletType]
					if list.ListProperties.NestingLevels[0].GlyphType == "DECIMAL" {
						sb.WriteString("1. ")
					} else {
						sb.WriteString("- ")
					}
				}
			}

			// Handle text runs
			for _, elem := range elem.Paragraph.Elements {
				if elem.TextRun != nil {
					text := elem.TextRun.Content
					if elem.TextRun.TextStyle != nil {
						text = applyTextStyle(text, elem.TextRun.TextStyle)
					}
					sb.WriteString(strings.TrimRight(text, " "))
				}
			}
			sb.WriteString("\n\n")
		}

		// Handle section breaks
		if elem.SectionBreak != nil {
			sb.WriteString("\n---\n")
		}

		// Handle tables
		if elem.Table != nil {
			sb.WriteString("\n")
			for rowIndex, row := range elem.Table.TableRows {
				sb.WriteString("|")
				for _, cell := range row.TableCells {
					sb.WriteString(" ")
					for _, cellElem := range cell.Content {
						if cellElem.Paragraph != nil {
							for _, cellElemElem := range cellElem.Paragraph.Elements {
								if cellElemElem.TextRun != nil {
									text := cellElemElem.TextRun.Content
									if cellElemElem.TextRun.TextStyle != nil {
										text = applyTextStyle(text, cellElemElem.TextRun.TextStyle)
									}
									sb.WriteString(strings.TrimRight(text, " "))
								}
							}
						}
					}
					sb.WriteString(" |")
				}
				sb.WriteString("\n")

				// Add table header separator after the first row
				if rowIndex == 0 {
					sb.WriteString("|")
					for range row.TableCells {
						sb.WriteString(" --- |")
					}
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	// Handle inline images
	imageCounter := 1
	for _, inlineObject := range doc.InlineObjects {
		if inlineObject.InlineObjectProperties != nil && inlineObject.InlineObjectProperties.EmbeddedObject != nil {
			if inlineObject.InlineObjectProperties.EmbeddedObject.ImageProperties != nil {
				imgSrc := inlineObject.InlineObjectProperties.EmbeddedObject.ImageProperties.ContentUri
				imageFileName := fmt.Sprintf("image%d.jpg", imageCounter)
				imageCounter++

				// Download the image
				imagePath := filepath.Join(docFolderPath, imageFileName)
				err := downloadImage(imgSrc, imagePath)
				if err != nil {
					return "", "", fmt.Errorf("unable to download image: %v", err)
				}

				// Insert image markdown in the content
				imageMarkdown := fmt.Sprintf("![Image](%s)\n\n", imageFileName)
				sb.WriteString(imageMarkdown)
			}
		}
	}

	// Consolidate formatting
	formattedText := ConsolidateMarkdownFormatting(sb.String())

	return formattedText, doc.Title, nil
}

// ConsolidateMarkdownFormatting ensures that markdown formatting symbols are on the same line and removes newlines and spaces before closing tags.
func ConsolidateMarkdownFormatting(text string) string {
	// Define regex patterns for different markdown formatting to consolidate newlines
	consolidatePatterns := []struct {
		pattern *regexp.Regexp
		repl    string
	}{
		{regexp.MustCompile(`\*\*([^\*]+?)\n([^\*]+?)\*\*`), "**$1$2**\n\n"},          // Bold
		{regexp.MustCompile(`_([^_]+?)\n([^_]+?)_`), "_$1$2_\n\n"},                    // Italics
		{regexp.MustCompile(`~~([^~]+?)\n([^~]+?)~~`), "~~$1$2~~\n\n"},                // Strikethrough
		{regexp.MustCompile(`<ins>([^<]+?)\n([^<]+?)<\/ins>`), "<ins>$1$2</ins>\n\n"}, // Underline
	}

	// Apply replacements to consolidate newlines within markdown tags
	for _, replacement := range consolidatePatterns {
		text = replacement.pattern.ReplaceAllString(text, replacement.repl)
	}

	// Remove spaces before closing tags
	cleanupPatterns := []struct {
		pattern *regexp.Regexp
		repl    string
	}{
		{regexp.MustCompile(` \*\*`), "**"},         // Bold
		{regexp.MustCompile(` _`), "_"},             // Italics
		{regexp.MustCompile(` ~~`), "~~"},           // Strikethrough
		{regexp.MustCompile(` </ins>`), "</ins>"},   // Underline
		{regexp.MustCompile(`\n\*\*`), "**"},        // Newline before bold closing
		{regexp.MustCompile(`\n_`), "_"},            // Newline before italics closing
		{regexp.MustCompile(`\n~~`), "~~"},          // Newline before strikethrough closing
		{regexp.MustCompile(`\n<\/ins>`), "</ins>"}, // Newline before underline closing
	}

	for _, cleanup := range cleanupPatterns {
		text = cleanup.pattern.ReplaceAllString(text, cleanup.repl)
	}

	return text
}

// applyTextStyle applies markdown styles based on the provided text style.
func applyTextStyle(text string, style *docs.TextStyle) string {
	styles := []string{}

	if style.Link != nil && style.Link.Url != "" {
		text = fmt.Sprintf("[%s](%s)", text, style.Link.Url)
	}

	if style.Bold {
		styles = append(styles, "**")
	}
	if style.Italic {
		styles = append(styles, "_")
	}
	if style.Underline {
		styles = append(styles, "<ins>")
	}
	if style.Strikethrough {
		styles = append(styles, "~~")
	}

	// Apply opening styles
	for _, s := range styles {
		text = s + text
	}

	// Apply closing styles in reverse order
	for i := len(styles) - 1; i >= 0; i-- {
		if styles[i] == "<ins>" {
			text = text + "</ins>"
		} else {
			text = text + strings.TrimRight(styles[i], " ")
		}
	}

	return text
}

// headingLevelToInt converts a heading level from string to int.
func headingLevelToInt(level string) int {
	switch level {
	case "HEADING_1":
		return 1
	case "HEADING_2":
		return 2
	case "HEADING_3":
		return 3
	case "HEADING_4":
		return 4
	case "HEADING_5":
		return 5
	case "HEADING_6":
		return 6
	default:
		return 1
	}
}

// GetRevisionContent retrieves the content of a specific revision of a Google Doc.
func GetRevisionContent(client *http.Client, docID, revisionID string) (string, error) {
	// Create the URL for the specific revision content
	downloadURL := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/export?mimeType=text/plain&revisionId=%s", docID, revisionID)
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create request for document revision: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve document revision: %v", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read revision content: %v", err)
	}
	return string(content), nil
}

// SaveMarkdownToFile saves markdown content to a specified file.
func SaveMarkdownToFile(content, folderName, filename string) error {
	os.MkdirAll(folderName, 0700)
	filePath := filepath.Join(folderName, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to create file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("unable to write to file: %v", err)
	}

	fmt.Printf("Markdown content saved to: %s\n", filePath)
	return nil
}

// downloadImage downloads an image from the specified URL and saves it to the specified path.
func downloadImage(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("unable to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read image data: %v", err)
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("unable to write image to file: %v", err)
	}

	return nil
}

package main

import (
	"fmt"
	"gdocs2md/internal/docs"
)

func main() {
	// Sample text with newlines within markdown formatting
	sampleText := `Test1
**Line1isbold**
_Line2isitalics_
<u>Line3isunderlined:</u>`

	// Consolidate markdown formatting
	consolidatedText := docs.ConsolidateMarkdownFormatting(sampleText)

	// Print the result
	fmt.Println(consolidatedText)
}

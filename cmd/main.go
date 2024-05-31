package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gdocs2md/internal/auth"
	docutils "gdocs2md/internal/docs"
	driveutils "gdocs2md/internal/drive"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func main() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		log.Fatalf("Environment variables GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URI must be set")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Endpoint:     google.Endpoint,
		Scopes:       []string{docs.DocumentsReadonlyScope, drive.DriveReadonlyScope, "https://www.googleapis.com/auth/drive.metadata.readonly"},
	}

	client := auth.GetClient(config)
	docsSrv, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	driveSrv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	var folderURL string
	fmt.Print("Enter the Google Drive folder URL: ")
	if _, err := fmt.Scan(&folderURL); err != nil {
		log.Fatalf("Unable to read folder URL: %v", err)
	}

	folderID := driveutils.ExtractFolderID(folderURL)
	if folderID == "" {
		log.Fatalf("Invalid folder URL")
	}

	folderName, err := driveutils.GetFolderName(driveSrv, folderID)
	if err != nil {
		log.Fatalf("Unable to retrieve folder name: %v", err)
	}

	files, err := driveutils.ListFilesInFolder(driveSrv, folderID)
	if err != nil {
		log.Fatalf("Unable to list files in folder: %v", err)
	}

	baseFolderPath := filepath.Join(".", folderName)
	err = os.MkdirAll(baseFolderPath, 0700)
	if err != nil {
		log.Fatalf("Unable to create base folder: %v", err)
	}

	for _, file := range files {
		docID := file.Id
		markdown, docTitle, err := docutils.ConvertDocsToMarkdown(docsSrv, docID, baseFolderPath)
		if err != nil {
			log.Fatalf("Unable to convert document to markdown: %v", err)
		}

		comments, err := driveutils.FetchComments(driveSrv, docID)
		if err != nil {
			log.Fatalf("Unable to fetch comments: %v", err)
		}

		if len(comments) > 0 {
			markdown += "\n\n## COMMENTS:\n"
			for i, comment := range comments {
				markdown += fmt.Sprintf("**Comment %d by %s (%s) on %s**: %s\n", i+1, comment.Author.DisplayName, comment.Author.EmailAddress, comment.CreatedTime, comment.Content)
			}
		}

		// Save the markdown content to a file in a folder named after the document title
		docFolderPath := filepath.Join(baseFolderPath, docTitle)
		err = docutils.SaveMarkdownToFile(markdown, docFolderPath, "document.md")
		if err != nil {
			log.Fatalf("Unable to save markdown to file: %v", err)
		}

		revisions, err := driveutils.FetchRevisions(driveSrv, docID)
		if err != nil {
			log.Fatalf("Unable to fetch revisions: %v", err)
		}

		for _, revision := range revisions {
			revisionMarkdown := fmt.Sprintf("**Revision ID:** %s\n**Modified Time:** %s\n**Modified By:** %s (%s)\n\n", revision.Id, revision.ModifiedTime, revision.LastModifyingUser.DisplayName, revision.LastModifyingUser.EmailAddress)
			revisionContent, err := docutils.GetRevisionContent(client, docID, revision.Id)
			if err != nil {
				log.Fatalf("Unable to retrieve revision content: %v", err)
			}

			revisionMarkdown += revisionContent
			revisionMarkdown = docutils.ConsolidateMarkdownFormatting(revisionMarkdown)

			// Create a filename with the timestamp
			timestamp := time.Now().Format("20060102T150405")
			revisionFilename := fmt.Sprintf("%s_%s.md", timestamp, revision.Id)
			err = docutils.SaveMarkdownToFile(revisionMarkdown, docFolderPath, revisionFilename)
			if err != nil {
				log.Fatalf("Unable to save revision markdown to file: %v", err)
			}
		}
	}
}

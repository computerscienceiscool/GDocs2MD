package drive

import (
	"fmt"
	"regexp"

	"google.golang.org/api/drive/v3"
)

// ListFilesInFolder lists all files in a given Google Drive folder.
func ListFilesInFolder(driveSrv *drive.Service, folderID string) ([]*drive.File, error) {
	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.document'", folderID)
	fileList, err := driveSrv.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list files in folder: %v", err)
	}
	return fileList.Files, nil
}

// GetFolderName retrieves the name of the Google Drive folder.
func GetFolderName(driveSrv *drive.Service, folderID string) (string, error) {
	folder, err := driveSrv.Files.Get(folderID).Fields("name").Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve folder name: %v", err)
	}
	return folder.Name, nil
}

// ExtractFolderID extracts the folder ID from the Google Drive folder URL.
func ExtractFolderID(url string) string {
	re := regexp.MustCompile(`^https://drive.google.com/drive/(u/\d+/)?folders/([a-zA-Z0-9_-]+)$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 2 {
		return matches[2]
	}
	return ""
}

// FetchComments retrieves comments from a Google Doc.
func FetchComments(driveSrv *drive.Service, fileID string) ([]*drive.Comment, error) {
	comments, err := driveSrv.Comments.List(fileID).Fields("comments(content,author(displayName,emailAddress),createdTime)").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve comments: %v", err)
	}
	return comments.Comments, nil
}

// FetchRevisions retrieves revisions from a Google Doc.
func FetchRevisions(driveSrv *drive.Service, fileID string) ([]*drive.Revision, error) {
	revisions, err := driveSrv.Revisions.List(fileID).Fields("revisions(id,modifiedTime,lastModifyingUser(displayName,emailAddress))").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve revisions: %v", err)
	}
	return revisions.Revisions, nil
}

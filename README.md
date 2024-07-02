
 
# GDocs2MD
GDocs2MD is a command-line tool that converts Google Docs documents within a specified Google Drive folder to Markdown format and saves them locally. It also fetches comments and revisions from the documents and saves them along with the main content.


## Prerequisites
-**Go**: Ensure you have Go installed. You can download it from golang.org.

-**Google API Credentials**: You'll need a Google Cloud project with the Google Drive and Google Docs APIs enabled. Create OAuth 2.0 credentials and download thecredentials.json file.

## Setup
**Environment Variables**
Set the following environment variables with the credentials from your Google Cloud project:


- GOOGLE_CLIENT_ID: Your OAuth 2.0 client ID.

- CLIENT_SECRET: Your OAuth 2.0 client secret.

- REDIRECT_URI: Your OAuth 2.0 redirect URI. This is typically set tohttp://localhost:8080.



## Install Dependencies
Run the following command to install the required dependencies:


go mod tidy


## Usage
- Run the Program ```go run cmd/main.go```


- Authorize the Application:


- The program will prompt you to open a URL in your browser.


- Open the URL, log in to your Google account, and grant the necessary permissions.


- Copy the authorization code provided and paste it back into the terminal.


- Enter the Google Drive Folder URL:


- After authentication, the program will prompt you to enter the URL of the Google Drive folder you want to process.


- Paste the folder URL into the terminal and press Enter.


Conversion Process:


- The program will retrieve the documents from the specified Google Drive folder.


- Each document will be converted to Markdown format, including comments and revisions.


- The converted files will be saved in a local folder named after the Google Drive folder.


## Expected Behavior
**Inputs**
- Authorization Code: After opening the URL provided by the program and authorizing access, you'll receive an authorization code. Enter this code into the terminal when prompted.


- Google Drive Folder URL: Enter the URL of the Google Drive folder containing the documents you want to convert.


**Outputs**
- The program will create a folder in the current directory named after the Google Drive folder.


- Inside this folder, you'll find:


- Markdown files for each Google Doc in the folder.


- Separate Markdown files for each revision of the documents.


- Comments included at the end of each document's Markdown file.


**File Structure**
- cmd/main.go: The main program file.


- internal/auth: Handles OAuth2 authentication.


- internal/docs: Contains functions for converting Google Docs to Markdown and saving files.


- internal/drive: Contains functions for interacting with Google Drive API.


**License**

This project is licensed under the MIT License. See the LICENSE file for details.





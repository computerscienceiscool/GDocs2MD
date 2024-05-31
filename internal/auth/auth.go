package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"golang.org/x/oauth2"
)

// GetClient retrieves an OAuth2 token, saves it if necessary, and returns the generated HTTP client.
func GetClient(config *oauth2.Config) *http.Client {
	tokFile := tokenCacheFile()
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	} else {
		tokSource := config.TokenSource(context.Background(), tok)
		tok, err = tokSource.Token()
		if err != nil {
			tok = getTokenFromWeb(config)
			saveToken(tokFile, tok)
		}
	}
	return config.Client(context.Background(), tok)
}

// startLocalServer starts a local web server to handle the OAuth2 callback.
func startLocalServer(config *oauth2.Config, state string, done chan<- string) {
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "State does not match", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in the response", http.StatusBadRequest)
			return
		}
		fmt.Fprintln(w, "Authorization successful, you can close this window.")
		done <- code
	})

	port := 8080
	log.Printf("Starting local web server on port %d", port)
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// getTokenFromWeb uses Config to request a Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	state := "state-token"
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)

	// Start local server to receive the authorization code
	done := make(chan string)
	go startLocalServer(config, state, done)

	// Open browser for user authorization
	err := openBrowser(authURL)
	if err != nil {
		log.Fatalf("Unable to open browser: %v", err)
	}

	// Wait for the authorization code from the local server
	authCode := <-done

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	saveToken(tokenCacheFile(), tok)
	return tok
}

// openBrowser opens the URL in the default browser.
func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

// tokenCacheFile generates credential file path/filename.
func tokenCacheFile() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to determine current user: %v", err)
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, "token.json")
}

// tokenFromFile retrieves a Token from a given file path.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken uses a file path to create a file and store the token in it.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

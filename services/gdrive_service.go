package services

import (
	context "context"
	json "encoding/json"
	fmt "fmt"
	ioutil "io/ioutil"
	log "log"
	http "net/http"
	os "os"
	user "os/user"
	filepath "path/filepath"

	oauth2 "golang.org/x/oauth2"
	google "golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	option "google.golang.org/api/option"
)

// UploadFileToGoogleDrive mengunggah file yang diberikan ke Google Drive
func UploadFileToGoogleDrive(filePath, fileName string) error {
	folderID := os.Getenv("GDRIVE_FOLDER_ID")
	ctx := context.Background()

	// Baca credentials dari file
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("unable to read credentials.json: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret: %v", err)
	}

	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fileMetadata := &drive.File{
		Name:    fileName,
		Parents: []string{folderID},
	}

	_, err = srv.Files.Create(fileMetadata).Media(f).Do()
	if err != nil {
		return fmt.Errorf("upload error: %v", err)
	}

	return nil
}

// Dapatkan klien HTTP terautentikasi dari credentials.json
func getClient(config *oauth2.Config) *http.Client {
	tokFile := tokenFilePath()
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFilePath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, ".credentials", "token.json")
}

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

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	os.MkdirAll(filepath.Dir(path), 0700)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code:\n%v\n", authURL)

	var authCode string
	fmt.Print("Enter verification code: ")
	fmt.Scan(&authCode)

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

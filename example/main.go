package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/Tobi696/googlesheetsparser"
)

type Workout struct {
	ID             uint
	NameLine1      string
	NameLine2      *string
	ImagePath      string
	Description    string
	Difficulty     uint
	Combustion     float64
	CombustionUnit string
	IsFree         bool

	CategoryID uint
}

type jwtConfig struct {
	Email        string   `json:"client_email"`
	PrivateKey   string   `json:"private_key"`
	PrivateKeyID string   `json:"private_key_id"`
	TokenURI     string   `json:"token_uri"`
	Scopes       []string `json:"scopes"`
}

// getService returns a Google Sheets API service
// using the credentials in "credentials.json"
// this code works for Service Accounts only
func getService(ctx context.Context) *sheets.Service {
	// Authenticating, creating the googlesheets Service
	var fileConf jwtConfig
	confFile, err := os.Open("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}
	defer confFile.Close()
	if err := json.NewDecoder(confFile).Decode(&fileConf); err != nil {
		log.Fatalf("Unable to parse credentials file: %v", err)
	}

	conf := &jwt.Config{
		Email:        fileConf.Email,
		PrivateKey:   []byte(fileConf.PrivateKey),
		PrivateKeyID: fileConf.PrivateKeyID,
		TokenURL:     fileConf.TokenURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets.readonly",
		},
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(conf.Client(ctx)))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}

func main() {
	ctx := context.Background()
	srv := getService(ctx)

	// Acutal usage of the Library
	users, err := googlesheetsparser.ParseSheetIntoStructSlice[Workout](ctx, googlesheetsparser.Options{
		Service:       srv,
		SpreadsheetID: "15PTbwnLdGJXb4kgLVVBtZ7HbK3QEj-olOxsY7XTzvCc",
		DatetimeFormats: []string{
			"2.1.2006",
			"02.01.2006",
			"02.01.2006 15:04:05",
		},
	}.Build())
	if err != nil {
		log.Fatalf("Unable to parse page: %v", err)
	}

	// Do anything you want with the users
	fmt.Println(users)
}

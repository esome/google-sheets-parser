package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/esome/google-sheets-parser"
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
	defer func() {
		_ = confFile.Close()
	}()
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

// spreadsheetID is the ID of the Google Sheet to be parsed. This is publicly accessible.
const spreadsheetID = "15PTbwnLdGJXb4kgLVVBtZ7HbK3QEj-olOxsY7XTzvCc"

func main() {
	ctx := context.Background()
	svc := getService(ctx)

	useOTC := flag.Bool("one-time-cfg", false, "Run the one-time-config example code")
	flag.Parse()

	if *useOTC {
		oneTimeConfig(ctx, svc)
		return
	}
	recommended(ctx, svc)
}

// recommended demonstrates the idiomatic usage of the Library
func recommended(ctx context.Context, svc *sheets.Service) {
	// Create a common Config with the Google Sheets service and the spreadsheet ID, and commonly used options
	// These options can still be overridden when calling the ParseSheetIntoStructSlice/ParseSheetIntoStructs functions
	// The config can still be reused for multiple calls. Options passed to the parsing functions will not taint the config.
	cfg := gsheets.MakeConfig(svc, spreadsheetID,
		gsheets.WithDatetimeFormats(
			"2.1.2006",
			"02.01.2006",
			"02.01.2006 15:04:05",
		))

	// Parse the sheet into a slice of structs
	users, err := gsheets.ParseSheetIntoStructSlice[Workout](ctx, cfg)
	if err != nil {
		log.Fatalf("Unable to parse page: %v", err)
	}

	// Do anything you want with the users
	fmt.Println(users)
}

// oneTimeConfig demonstrates the idiomatic usage of the Library
func oneTimeConfig(ctx context.Context, svc *sheets.Service) {
	// Parse the sheet into a slice of structs
	users, err := gsheets.ParseSheetIntoStructSlice[Workout](ctx, gsheets.Config{Service: svc},
		gsheets.WithSpreadsheetID(spreadsheetID),
		gsheets.WithDatetimeFormats(
			"2.1.2006",
			"02.01.2006",
			"02.01.2006 15:04:05",
		))
	if err != nil {
		log.Fatalf("Unable to parse page: %v", err)
	}

	// Do anything you want with the users
	fmt.Println(users)
}

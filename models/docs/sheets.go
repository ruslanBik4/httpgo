// Copyright 2017 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package docs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"

	"github.com/ruslanBik4/httpgo/logs"
	"github.com/ruslanBik4/httpgo/models/server"
)

const ClientID = "165049723351-mgcbnem17vt14plfhtbfdcerc1ona2p7.apps.googleusercontent.com"
const authCode = "4/H7iL6R6BSstU5-W0V7WgI9cPZttAjOzHH5pEmwYS8UQ#"

type SheetsGoogleDocs struct {
	Service *sheets.Service
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), err
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	//authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Printf("Go to the following link in your browser then type the "+
	//	"authorization code: \n%v\n", authURL)

	//var code string
	//if _, err := fmt.Scan(&code); err != nil {
	//	log.Println("Unable to read authorization code %v", err)
	//}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		logs.ErrorLog(errors.New("Unable to retrieve token from web"), err)
	}
	return tok
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logs.ErrorLog(errors.New("Unable to cache oauth token: "), err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

var cacheFile = "config/.credentials"
var userFile = "config/oauth2.json"

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	//cacheFile, err := tokenCacheFile()
	//if err != nil {
	//	log.Println("Unable to get path to cached credential file. %v", err)
	//}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		logs.ErrorLog(err, cacheFile)
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}
func newClient() *http.Client {
	ctx := context.Background()
	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	b, err := ioutil.ReadFile(userFile)
	if err != nil {
		logs.ErrorLog(errors.New("Unable to read client secret file: "), err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		logs.ErrorLog(errors.New("Unable to parse client secret file to config: "), err)
	}
	return getClient(ctx, config)
}
func (sheet *SheetsGoogleDocs) Init() (err error) {

	sConfig := server.GetServerConfig()
	cacheFile = filepath.Join(sConfig.SystemPath(), "config/.credentials")
	userFile = filepath.Join(sConfig.SystemPath(), "config/oauth2.json")

	if sheet.Service, err = sheets.New(newClient()); err != nil {
		return err
	}

	return nil
}

func (sheet *SheetsGoogleDocs) Read(spreadsheetId, readRange string) (*sheets.ValueRange, error) {

	resp, err := sheet.Service.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}
func (sheet *SheetsGoogleDocs) Sheets(spreadsheetId, readRange string) (*sheets.ValueRange, error) {
	sh := sheet.Service.Spreadsheets.Values

	resp, err := sh.Get(spreadsheetId, readRange).Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

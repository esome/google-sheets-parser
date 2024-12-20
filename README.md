# Google Sheets Parser

[![license](https://img.shields.io/github/license/esome/google-sheets-parser?style=flat&label=License&labelColor=rgb(45%2C%2049%2C%2054)&color=rgb(113%2C%2016%2C%20126))](LICENSE.md) 
[![release](https://img.shields.io/github/v/release/esome/google-sheets-parser?include_prereleases&sort=date&display_name=release&style=flat&label=Release&labelColor=rgb(45%2C%2049%2C%2054))](https://github.com/esome/google-sheets-parser/releases) 
[![badge](https://github.com/esome/google-sheets-parser/workflows/CodeQL/badge.svg)](https://github.com/esome/google-sheets-parser/actions/workflows/github-code-scanning/codeql)
[![badge](https://github.com/esome/google-sheets-parser/workflows/Go/badge.svg)](https://github.com/esome/google-sheets-parser/actions/workflows/go.yml)
![badge](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/sGy1980de/b272dbf4526c9be75f7da96352873a71/raw/gsheets-parser-coverage.json)

Google Sheets Parser is a library for dynamically parsing Google Sheets into Golang structs.

## Installation

```shell
go get github.com/esome/google-sheets-parser
```

### Requirements

This library requires Go >= 1.23 as range over function mechanics are used.

## Usage

![Example Sheet](Users_Sheet.png)
The Image shows the sheet called "Users" which is contained in the example spreadsheet.  

To Parse it, we would utilize following code:

```go
package main

import (
	"fmt"
	"log"
	"time"
	
	"github.com/esome/google-sheets-parser"
	"google.golang.org/api/sheets/v4"
)

// User is a struct that represents a row in the Users Sheet
type User struct {
	ID        uint // <- By default, columns will be parsed into the equally named struct fields.
	Username  string
	Name      string
	Password  *string
	Weight    *uint
	CreatedAt *time.Time `gsheets:"Created At"` // <- Custom Column Name, optional, will be prioritized over the Struct Field Name
}

func main() {
	var svc *sheets.Service // <- You need to create a Google Sheets Service first, see below
	
	users, err := gsheets.ParseSheetIntoStructSlice[User](
		// minimal Config only containing the Google Sheets service (*sheets.Service)
		// Have a look in the example to learn how to create a reusable configuration 
		gsheets.Config{Service: svc},
		// Mandatory! you must define the SpreadsheetID, or an error will be returned
		gsheets.WithSpreadsheetID("15PTbwnLdGJXb4kgLVVBtZ7HbK3QEj-olOxsY7XTzvCc"),
		// Optional: you can pass an arbitrary amount of ConfigOptions for further customization for this call
		gsheets.WithDatetimeFormats( // <- in this case we provide further Datetime Formats to be recognized 
			"2.1.2006",
			"02.01.2006",
			"02.01.2006 15:04:05",
		),
	)
	if err != nil {
		log.Fatalf("Unable to parse page: %v", err)
	}

	// Do anything you want with the result
	fmt.Println(users)
}
```


### Authenticating a Google Sheets Service

There are different ways to authenticate a Google Sheets Service.

Please refer to the [Google Sheets Go Quickstart](https://developers.google.com/sheets/api/quickstart/go) for more information.


### Example

To try out the example yourself, check out the [example/](example/)-Directory.  
In there you will find an example that demonstrates how to create a common config for multiple parse calls.


## Intention

This library is intended to be used as a library for parsing Google Sheets into Golang structs. It is not intended to be used as a library for generating Google Sheets from Golang structs.  

At esome we use Google Sheets in some cases to communicate data with external partners. The data of these sheets sometimes
needs to be imported into our data-warehouse. This library helps us to parse the data from the Google Sheets into 
Golang structs, which then can be written to the databases.  


## Origin

This library was originally a fork of the [awesome work from Tobias Wimmer](https://github.com/Tobi696/googlesheetsparser).
All credits go to him for the initial implementation. Please consider checking out his repository as well, and support him.

Since the way the library is configured has changed significantly, we decided to create a complete separate/independent
repository for it. This also allows us to maintain the library in a way that fits our needs better, and accept/make
contributions that may have been rejected in the original project. 


## Contributing

Contributions are welcome! Please open an issue or pull request if you have any suggestions or want to contribute.


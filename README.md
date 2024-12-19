# Google Sheets Parser

`google-sheets-parser` is a library for dynamically parsing Google Sheets into Golang structs.

## Installation

```bash
go get github.com/esome/google-sheets-parser
```

### Requirements

This library requires Go >= 1.23 as range over function mechanics are used.

## Usage

![Example Sheet](Users_Sheet.png)
The Image shows the sheet called "Users" which is contained in the example spreadsheet.  

To Parse it, we would utilize following code:

```go
// Define your structs to be parsed
type User struct {
	ID        uint // <- By default, columns will be parsed into the equally named struct fields.
	Username  string
	Name      string
	Password  *string
	Weight    *uint
	CreatedAt *time.Time `gsheets:"Created At"` // <- Custom Column Name, optional, will be prioritized over the Struct Field Name
}

// Acutal usage of the Library
users, err := gsheets.ParseSheetIntoStructSlice[User](
	context.Background(), 
	// minimal Config only containing the Google Sheets service (*sheets.Service)
	gsheets.Config{Service: svc},
	// Mandatory! you must define the SpreadsheetID, or an error will be returned
	gsheets.WithSpreadsheetID( "15PTbwnLdGJXb4kgLVVBtZ7HbK3QEj-olOxsY7XTzvCc"),
	// Optional: you can pass an arbitrary amount of ConfigOptions for further customization for this call
    gsheets.WithDatetimeFormats:() // <- in this case we provide further Datetime Formats to be recognized 
        "2.1.2006",
        "02.01.2006",
        "02.01.2006 15:04:05",
    },
) 
if err != nil {
    log.Fatalf("Unable to parse page: %v", err)
}

// Do anything you want with the result
fmt.Println(users)
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


## Contributing

Contributions are welcome! Please open an issue or pull request if you have any suggestions or want to contribute.


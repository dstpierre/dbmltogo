# dbmltogo
Generate Go and Elm(preview) data access code from a C# .NET DBML file.

## Motivation

I work as a consultant in a >25 years old fintech company; way before the word 
fintech existed. There's some legacy C# ASP.NET applications that use 
**Linq-to-SQL** (read not Entity Framework) that's not supported anymore in .NET 
Core. 

Instead of migrating those web apps to .NET Core and rewrites those parts, I'm 
restarting fresh new projects in Go and Elm.

The goal of this tool is to generate the initial boilerplate code needed on the 
backend and frontend for all CRUD operations.

## Usage

```shell
$> git clone git@github.com:dstpierre/dbmltogo.git && cd dbmltogo
$> go build
$> ./dbmltogo -dbml your-file.dbml -pkgname data -elm -prefix prefix-for-your-url
```

**-dbml**: this is your DBML file. This was tested only with C# projects.

**-pkgname**: This is your Go package name for the generated data file, (default to data).

**-elm**: Optional, will generate Elm module for all `<Table>` in the DBML file.

**-prefix**: Optional, when generating Elm module this is used on the endpoints URL i.e. /[prefix]/banks 
(default to empty).

This only works for Microsoft SQL Server.

## What it is generating

### Go files

It will create one Go file for all `<Table>` in a **DBML** file:

Example: *./[pkgname]/[member].go* -> **./data/banks.go**

```go
package data

import (
	"database/sql"
)

// Bank ...
type Bank struct {
	BankID         int            `json:"bankId"`
	BankNameFr     JSONNullString `json:"bankNameFr"`
	BankNameEn     JSONNullString `json:"bankNameEn"`
	CentralAddress JSONNullString `json:"centralAddress"`
	CentralPhone   JSONNullString `json:"centralPhone"`
	CentralFax     JSONNullString `json:"centralFax"`
	CommentA       JSONNullString `json:"commentA"`
	CommentB       JSONNullString `json:"commentB"`
	Short          JSONNullString `json:"short"`
	Fee            float64        `json:"fee"`
	DisplayOrder   JSONNullInt32  `json:"displayOrder"`
}

// Banks ...
type Banks struct {
	DB *sql.DB
}

// Create creates a new Bank
func (repo *Banks) Create(entity Bank) (id int, err error) {
	err = repo.DB.QueryRow(`
		INSERT INTO dbo.Bank
		VALUES(@BankNameFr,
		@BankNameEn,
		@CentralAddress,
		@CentralPhone,
		@CentralFax,
		@CommentA,
		@CommentB,
		@Short,
		@Fee,
		@DisplayOrder,
		
		)
		SELECT SCOPE_IDENTITY()
	`, sql.Named("BankNameFr", entity.BankNameFr),
		sql.Named("BankNameEn", entity.BankNameEn),
		sql.Named("CentralAddress", entity.CentralAddress),
		sql.Named("CentralPhone", entity.CentralPhone),
		sql.Named("CentralFax", entity.CentralFax),
		sql.Named("CommentA", entity.CommentA),
		sql.Named("CommentB", entity.CommentB),
		sql.Named("Short", entity.Short),
		sql.Named("Fee", entity.Fee),
		sql.Named("DisplayOrder", entity.DisplayOrder),
	).Scan(&id)
	return
}

// List lists all Bank
func (repo *Banks) List() ([]Bank, error) {
	rows, err := repo.DB.Query(`
		SELECT
			BankID,
			BankNameFr,
			BankNameEn,
			CentralAddress,
			CentralPhone,
			CentralFax,
			CommentA,
			CommentB,
			Short,
			Fee,
			DisplayOrder,
			
		FROM dbo.Bank
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []Bank{}
	for rows.Next() {
		entity := Bank{}
		if err := repo.scan(rows, &entity); err != nil {
			return nil, err
		}
		results = append(results, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByID gets a Bank by BankID
func (repo *Banks) GetByID(id int) (*Bank, error) {
	row := repo.DB.QueryRow(`
		SELECT
			BankID,
			BankNameFr,
			BankNameEn,
			CentralAddress,
			CentralPhone,
			CentralFax,
			CommentA,
			CommentB,
			Short,
			Fee,
			DisplayOrder,
			
		FROM dbo.Bank
		WHERE BankID = @id
	`, sql.Named("id", id))

	entity := &Bank{}
	if err := repo.scan(row, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// Update saves a Bank
func (repo *Banks) Update(id int, entity Bank) error {
	_, err := repo.DB.Exec(`
		UPDATE dbo.Bank SET
			BankNameFr = @BankNameFr,
			BankNameEn = @BankNameEn,
			CentralAddress = @CentralAddress,
			CentralPhone = @CentralPhone,
			CentralFax = @CentralFax,
			CommentA = @CommentA,
			CommentB = @CommentB,
			Short = @Short,
			Fee = @Fee,
			DisplayOrder = @DisplayOrder,
			
		WHERE BankID = @BankID
	`, sql.Named("BankID", entity.BankID),
		sql.Named("BankNameFr", entity.BankNameFr),
		sql.Named("BankNameEn", entity.BankNameEn),
		sql.Named("CentralAddress", entity.CentralAddress),
		sql.Named("CentralPhone", entity.CentralPhone),
		sql.Named("CentralFax", entity.CentralFax),
		sql.Named("CommentA", entity.CommentA),
		sql.Named("CommentB", entity.CommentB),
		sql.Named("Short", entity.Short),
		sql.Named("Fee", entity.Fee),
		sql.Named("DisplayOrder", entity.DisplayOrder),
	)
	return err
}

// Delete delets a Bank by BankID
func (repo *Banks) Delete(id int) error {
	_, err := repo.DB.Exec(`
		DELETE FROM dbo.Bank WHERE BankID = @id
	`, sql.Named("id", id))
	return err
}

func (repo *Banks) scan(rows scanner, entity *Bank) error {
	return rows.Scan(&entity.BankID,
		&entity.BankNameFr,
		&entity.BankNameEn,
		&entity.CentralAddress,
		&entity.CentralPhone,
		&entity.CentralFax,
		&entity.CommentA,
		&entity.CommentB,
		&entity.Short,
		&entity.Fee,
		&entity.DisplayOrder,
	)
}

```

#### Description

**Plural name**: is the type that gets all the data function attached. You would 
use it like this:

```go
banksRepo := &data.Banks{DB: db}
b, err := banksRepo.GetByID(123)
//...
```

You need to pass the `*sql.DB`.

**Singular name**: is the struct with all the fields. Notice that nullable fields 
use a custom JSON types that will encode to `null` when null instead of the 
`{"valid": false, "string": ""}`'s `sql.NullX` types.

**There's extra ,**: some queries have leading "," and will fail when executing 
without manual modification.

**scanner type**: this is useful to scan from both `sql.Rows` and `sql.Row`.

**One primary key**: your tables need to have 1 primary key. If there's combine 
keys only the first one will be used, so be careful if you have multi-columns PKs.

**gofmt and goimports**: both are executed against the generated Go files.



### Elm files

*Not as easy as the Go generated code for now*.

It will optionally generate Elm modules and Api endpoint files for all `<Table>` 
in the **DBML** file:

*Please note*: there's lots of Modules and functions taken from my Elm course: 
[Pragmatic Elm web apps](https://dominicstpierre.com/videos/pragmatic-elm-web-apps) 
at first this tool was going to focus only on Go, hence the name, but I needed 
to get Elm modules so I added an optional flag. And yes if you'd like to support 
me buying my course is very helpful.

*/elm/Data/[Entity].elm* -> **./elm/Data/Bank.elm**

```elm
module Data.Bank exposing
    ( Bank
    , create
    , destroy
    , getById
    , list
    , update
    )

import Api.Banks as Endpoints
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (required, optional)
import Json.Encode as Encode
import Json.Encode.Extra as Extra
import Session exposing (Session, delete, get, post, put)
import Time

-- TYPES

type alias Bank =
    { bankId : Int
    , bankNameFr : Maybe String
    , bankNameEn : Maybe String
    , centralAddress : Maybe String
    , centralPhone : Maybe String
    , centralFax : Maybe String
    , commentA : Maybe String
    , commentB : Maybe String
    , short : Maybe String
    , fee : Float
    , displayOrder : Maybe Int
    
    }


-- ENCODERS/DECODERS

decoder : Decoder Bank
decoder =
    Decode.succeed Bank
        |> required "bankId" Decode.int 
        |> optional "bankNameFr" Decode.string Nothing
        |> optional "bankNameEn" Decode.string Nothing
        |> optional "centralAddress" Decode.string Nothing
        |> optional "centralPhone" Decode.string Nothing
        |> optional "centralFax" Decode.string Nothing
        |> optional "commentA" Decode.string Nothing
        |> optional "commentB" Decode.string Nothing
        |> optional "short" Decode.string Nothing
        |> required "fee" Decode.float 
        |> optional "displayOrder" Decode.int Nothing
        


encode : Bank -> Encode.Value
encode x =
    Encode.object
        [ ("bankId", Encode.int x.bankId)
        , ("bankNameFr", Extra.maybe Encode.string x.bankNameFr)
        , ("bankNameEn", Extra.maybe Encode.string x.bankNameEn)
        , ("centralAddress", Extra.maybe Encode.string x.centralAddress)
        , ("centralPhone", Extra.maybe Encode.string x.centralPhone)
        , ("centralFax", Extra.maybe Encode.string x.centralFax)
        , ("commentA", Extra.maybe Encode.string x.commentA)
        , ("commentB", Extra.maybe Encode.string x.commentB)
        , ("short", Extra.maybe Encode.string x.short)
        , ("fee", Encode.float x.fee)
        , ("displayOrder", Extra.maybe Encode.int x.displayOrder)
        
        ]


-- API ACCESS

create : Session -> Bank -> (Result Http.Error Int -> msg) -> Cmd msg
create session x msg =
    post
        session
        Endpoints.create
        (Http.jsonBody (encode x))
        (Http.expectJson msg Decode.int)

list : Session -> (Result Http.Error (List Bank) -> msg) -> Cmd msg
list session msg =
    get
        session
        Endpoints.list
        (Http.expectJson msg (Decode.list decoder))

getById : Session -> Int -> (Result Http.Error Bank -> msg) -> Cmd msg
getById session id msg =
    get
        session
        (Endpoints.getById id)
        (Http.expectJson msg decoder)

update : Session -> (Int, Bank) -> (Result Http.Error Bool -> msg) -> Cmd msg
update session (id, x) msg =
    put
        session
        (Endpoints.update id)
        (Http.jsonBody (encode x))
        (Http.expectJson msg Decode.bool)

destroy : Session -> Int -> (Result Http.Error Bool -> msg) -> Cmd msg
destroy session id msg =
    delete
        session
        (Endpoints.delete id)
        (Http.expectJson msg Decode.bool)

```

#### Description

**Api endpoint**: the `Api.Endpoint` are generated those are helper functions to 
call your backend API.

**Nullable columns**: all nullable columns use `Maybe` and with the proper JSON 
encoder/decoder.

### Contribution

Feel free to submit apull request. Here's some todos:

1. Finding a way to remove the extra leading "," on SQL queries.
2. Handle <Function> in the DBML (stored procedure).
3. Handle <Association> and create structs and INNER JOIN queries.

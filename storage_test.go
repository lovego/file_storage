package filestorage

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

var testDB = getDB()
var testStorage = getStorage()
var testFileHeaders = getFileHeaders()

func ExampleStorage_Upload() {
	db, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	if files, err := testStorage.Upload(db, nil, "", testFileHeaders...); err != nil {
		panic(err)
	} else {
		fmt.Println(files)
	}
	if err := db.Commit(); err != nil {
		panic(err)
	}

	// Output:
	// [TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF0 HEbi8PV2cRPf8QeB8lesh6gWPAmiAda7xGarbjAv8v4]

}

func ExampleStorage_Link() {
	fmt.Println(testStorage.UnlinkAllOf(testDB, "object"))
	fmt.Println(testStorage.Link(testDB, "object", "file1", "file2", "file3"))
	fmt.Println(testStorage.FilesOf(testDB, "object"))
	// Output:
	// <nil>
	// <nil>
	// [file1 file2 file3] <nil>
}

func ExampleStorage_LinkOnly() {
	fmt.Println(testStorage.LinkOnly(testDB, "object", "file3", "file4"))
	fmt.Println(testStorage.FilesOf(testDB, "object"))
	fmt.Println(testStorage.EnsureLinked(testDB, "object", "file3"))
	fmt.Println(testStorage.Unlink(testDB, "object", "file3", "file4"))
	fmt.Println(testStorage.Linked(testDB, "object", "file3"))
	// Output:
	// <nil>
	// [file3 file4] <nil>
	// <nil>
	// <nil>
	// false <nil>
}

func getFileHeaders() []*multipart.FileHeader {
	form, err := multipart.NewReader(strings.NewReader(`
--ZnGpDtePMx0KrHh_G0X99Yef9r8JZsRJSXC
Content-Disposition: form-data;name="file"; filename="1.jpg"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

1.jpg
--ZnGpDtePMx0KrHh_G0X99Yef9r8JZsRJSXC
Content-Disposition: form-data;name="file"; filename="2.jpg"
Content-Type: application/octet-stream
Content-Transfer-Encoding: binary

2.txt
--ZnGpDtePMx0KrHh_G0X99Yef9r8JZsRJSXC--
`), "ZnGpDtePMx0KrHh_G0X99Yef9r8JZsRJSXC").ReadForm(1024)
	if err != nil {
		panic(err)
	}
	return form.File["file"]
}

func getStorage() *Storage {
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		panic(err)
	}
	s := Storage{
		ScpMachines: []string{"localhost"},
		ScpPath:     tmpDir,
	}
	if err := s.Init(testDB); err != nil {
		panic(err)
	}
	return &s
}

func getDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if _, err := db.Exec(`
		DROP TABLE IF EXISTS files;
		DROP TABLE IF EXISTS file_links;
	`); err != nil {
		panic(err)
	}
	return db
}

package filestorage

import (
	"context"
	"database/sql"
	"encoding/json"
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

const testFile1 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF1"
const testFile2 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF2"
const testFile3 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3"
const testFile4 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF4"

func ExampleStorage_Link() {
	fmt.Println(testStorage.UnlinkAllOf(testDB, "object"))
	fmt.Println(testStorage.Link(testDB, "object", testFile1, testFile2, testFile3))
	if files, err := testStorage.FilesOf(testDB, "object"); err != nil {
		fmt.Println(err)
	} else {
		for _, v := range files {
			fmt.Println(v)
		}
	}
	// Output:
	// <nil>
	// <nil>
	// TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF1
	// TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF2
	// TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3
}

func ExampleStorage_LinkOnly() {
	fmt.Println(testStorage.LinkOnly(testDB, "object", testFile3, testFile4))
	if files, err := testStorage.FilesOf(testDB, "object"); err != nil {
		fmt.Println(err)
	} else {
		for _, v := range files {
			fmt.Println(v)
		}
	}

	fmt.Println(testStorage.EnsureLinked(testDB, "object", testFile3))
	fmt.Println(testStorage.Unlink(testDB, "object", testFile3, testFile4))
	fmt.Println(testStorage.Linked(testDB, "object", testFile3))
	// Output:
	// <nil>
	// TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3
	// TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF4
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

func ExampleLinkObject_MarshalJSON() {
	object := LinkObject{Table: "x", ID: 1}
	b, err := json.Marshal(object)
	fmt.Println(string(b), err)
	// Output:
	// "x.1" <nil>
}

func ExampleLinkObject_UnmarshalJSON() {
	object := LinkObject{}
	err := json.Unmarshal([]byte(`"x.1.b"`), &object)
	fmt.Printf("%#v\n", object)
	fmt.Println(err)
	// Output:
	// filestorage.LinkObject{Table:"x", ID:1, Field:"b"}
	// <nil>
}

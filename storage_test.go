package file_storage

import (
	"database/sql"
	"fmt"
	"mime/multipart"
	"strings"

	_ "github.com/lib/pq"
)

var testDB = getDB()
var testFileHeaders = getFileHeaders()

func ExampleStorage_Upload() {
	s := Storage{
		ScpMachines: []string{"localhost"},
		ScpPath:     "tmp/",
	}
	if err := s.Init(testDB); err != nil {
		panic(err)
	}
	fmt.Println(s.Upload(testDB, nil, "", testFileHeaders...))

	// Output:
	// [TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF0 HEbi8PV2cRPf8QeB8lesh6gWPAmiAda7xGarbjAv8v4] <nil>
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

func getDB() *sql.DB {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}
	return db
}

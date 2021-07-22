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
	// [4c468b3b16999fd9578189576d5f770cb4a16ad9fca0e798a251f00a54a87c5d 1c46e2f0f5767113dff10781f257ac87a8163c09a201d6bbc466ab6e302ff2fe] <nil>
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

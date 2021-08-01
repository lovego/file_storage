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
var testBucket = getBucket()
var testFileHeaders = getFileHeaders()

func ExampleBucket_Upload() {
	db, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	if files, err := testBucket.Upload(db, nil, "", testFileHeaders...); err != nil {
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

func ExampleBucket_Link() {
	fmt.Println(testBucket.UnlinkAllOf(nil, "object"))
	fmt.Println(testBucket.Link(nil, "object", testFile1, testFile2, testFile3))
	if files, err := testBucket.FilesOf(nil, "object"); err != nil {
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

func ExampleBucket_LinkOnly() {
	fmt.Println(testBucket.LinkOnly(nil, "object", testFile3, testFile4))
	if files, err := testBucket.FilesOf(nil, "object"); err != nil {
		fmt.Println(err)
	} else {
		for _, v := range files {
			fmt.Println(v)
		}
	}

	fmt.Println(testBucket.EnsureLinked(nil, "object", testFile3))
	fmt.Println(testBucket.Unlink(nil, "object", testFile3, testFile4))
	fmt.Println(testBucket.Linked(nil, "object", testFile3))
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

func getBucket() *Bucket {
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		panic(err)
	}
	b := Bucket{
		Machines: []string{"localhost"},
		Dir:      tmpDir,
		DB:       testDB,
	}
	if err := b.Init(nil); err != nil {
		panic(err)
	}
	return &b
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
	// "x|1" <nil>
}

func ExampleLinkObject_UnmarshalJSON() {
	object := LinkObject{}
	err := json.Unmarshal([]byte(`"a.x|1|b"`), &object)
	fmt.Println(err)
	fmt.Printf("%#v\n", object)
	// Output:
	// <nil>
	// filestorage.LinkObject{Table:"a.x", ID:1, Field:"b"}
}

func ExampleLinkObject_UnmarshalJSON_2() {
	object := LinkObject{}
	err := object.UnmarshalJSON([]byte(`a.x|1|b`))
	fmt.Println(err)
	fmt.Printf("%#v\n", object)
	// Output:
	// <nil>
	// filestorage.LinkObject{Table:"a.x", ID:1, Field:"b"}
}

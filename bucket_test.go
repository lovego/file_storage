package filestorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/lovego/logger"
)

var testDB = getTestDB()
var testBucket = getTestBucket()
var testFileHeaders = getTestFileHeaders()

func ExampleBucket_Upload() {
	fmt.Println(testUpload())
	// Output:
	// [TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF0 HEbi8PV2cRPf8QeB8lesh6gWPAmiAda7xGarbjAv8v4]
}

func testUpload() []string {
	db, err := testDB.BeginTx(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	files, err := testBucket.Upload(db, nil, "", testFileHeaders...)
	if err != nil {
		panic(err)
	}
	if err := db.Commit(); err != nil {
		panic(err)
	}
	return files
}

func ExampleBucket_SaveFiles() {
	files, err := testBucket.SaveFiles(nil, nil, "linkObject", "LICENSE")
	if err != nil {
		panic(err)
	}
	fmt.Println(files)
	// Output:
	// [y_-6r_79lAd8cpzmKuK-W0u7_ZulcxaquXsi308-mqk]
}

const testFile1 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF1"
const testFile2 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF2"
const testFile3 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3"
const testFile4 = "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF4"

func ExampleBucket_Link() {
	testBucket.insertFileRecords(nil, []fileRecord{
		{Hash: "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF1"},
		{Hash: "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF2"},
		{Hash: "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3"},
	})
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
	testBucket.insertFileRecords(nil, []fileRecord{
		{Hash: "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF3"},
		{Hash: "TEaLOxaZn9lXgYlXbV93DLShatn8oOeYolHwClSofF4"},
	})
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

func ExampleBucket_StartClean() {
	testUpload()
	testBucket.StartClean(time.Second, time.Nanosecond, logger.New(os.Stdout))
	time.Sleep(time.Second)
	// Output:
}

func getTestFileHeaders() []*multipart.FileHeader {
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

func getTestBucket() *Bucket {
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

func getTestDB() *sql.DB {
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

func ExampleGetContentHash() {
	if os.Getenv("file") == "" {
		return
	}
	f, err := os.Open(os.Getenv("file"))
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	fmt.Println(getContentHash(f))

	// Output:
}

func ExampleImageChecker() {
	fmt.Println(imageChecker{"zh"}.fileSizeError(3123456))
	// Output: args-err: 文件大小(3,123,456)不能超过2兆.
}

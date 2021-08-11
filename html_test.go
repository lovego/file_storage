package filestorage

import "fmt"

func ExampleReplaceImgSrc() {
	var html = `<img> <img/>
<img src="1">
<img src="2"/>
<imG src="3" />
<IMG Src="4" / >
<img a src="5" b>
<img a src="x"b>
<img +src="x">
<img+ src="x">
`
	fmt.Println(ReplaceImgSrc(html, func(src string) string { return src + "~" }))
	// Output:
	// <img> <img/>
	// <img src="1~">
	// <img src="2~"/>
	// <imG src="3~" />
	// <IMG Src="4~" / >
	// <img a src="5~" b>
	// <img a src="x"b>
	// <img +src="x">
	// <img+ src="x">
}

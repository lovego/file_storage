package filestorage

import (
	"regexp"
	"strings"
)

const blankOrAttrs = `(?:\s[^<>]*)?`

var imgSrcRegexp = regexp.MustCompile(
	`(?i:<img` + blankOrAttrs + `\ssrc="(.*?)"` + blankOrAttrs + `/?>)`,
)

func ImgSrcToFileHash(html string) (result string, hashes []string) {
	return ReplaceImgSrc(html, func(src string) string {
		hash, _ := TryFileHash(src)
		if hash != "" && IsHash(hash) {
			hashes = append(hashes, hash)
		}
		return hash
	}), hashes
}

func (b *Bucket) ImgSrcToDownloadURL(linkObject interface{}, html string) string {
	return ReplaceImgSrc(html, func(src string) string {
		return b.DownloadURL(linkObject, src)
	})
}

func ReplaceImgSrc(html string, fn func(src string) string) string {
	indexes := imgSrcRegexp.FindAllStringSubmatchIndex(html, -1)
	if len(indexes) == 0 {
		return html
	}
	var newHtml = make([]string, len(indexes)*2+1)
	newHtml[0] = html[0:indexes[0][2]]
	for i := 0; i < len(indexes); i++ {
		newHtml[2*i+1] = fn(html[indexes[i][2]:indexes[i][3]])
		if i+1 < len(indexes) {
			newHtml[2*i+2] = html[indexes[i][3]:indexes[i+1][2]]
		} else {
			newHtml[2*i+2] = html[indexes[i][3]:]
		}
	}
	return strings.Join(newHtml, "")
}

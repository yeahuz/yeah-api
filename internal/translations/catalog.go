// Code generated by running "go generate" in golang.org/x/text. DO NOT EDIT.

package translations

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

type dictionary struct {
	index []uint32
	data  string
}

func (d *dictionary) Lookup(key string) (data string, ok bool) {
	p, ok := messageKeyToIndex[key]
	if !ok {
		return "", false
	}
	start, end := d.index[p], d.index[p+1]
	if start == end {
		return "", false
	}
	return d.data[start:end], true
}

func init() {
	dict := map[string]catalog.Dictionary{
		"en": &dictionary{index: enIndex, data: enData},
		"ru": &dictionary{index: ruIndex, data: ruData},
		"uz": &dictionary{index: uzIndex, data: uzData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"Welcome!\n": 0,
}

var enIndex = []uint32{ // 2 elements
	0x00000000, 0x0000000e,
} // Size: 32 bytes

const enData string = "\x04\x00\x01\n\t\x02Welcome!"

var ruIndex = []uint32{ // 2 elements
	0x00000000, 0x00000025,
} // Size: 32 bytes

const ruData string = "\x04\x00\x01\n \x02Добро пожаловать"

var uzIndex = []uint32{ // 2 elements
	0x00000000, 0x00000013,
} // Size: 32 bytes

const uzData string = "\x04\x00\x01\n\x0e\x02Hush kelibsiz"

// Total table size 166 bytes (0KiB); checksum: 6AE4A54C

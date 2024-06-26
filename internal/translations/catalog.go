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
		"de": &dictionary{index: deIndex, data: deData},
		"en": &dictionary{index: enIndex, data: enData},
		"fr": &dictionary{index: frIndex, data: frData},
	}
	fallback := language.MustParse("en")
	cat, err := catalog.NewFromMap(dict, catalog.Fallback(fallback))
	if err != nil {
		panic(err)
	}
	message.DefaultCatalog = cat
}

var messageKeyToIndex = map[string]int{
	"About":                   0,
	"App Preferences":         8,
	"App Registration":        5,
	"App Settings":            2,
	"Auto-discovered Servers": 13,
	"Fyne Preferences":        7,
	"Fyne Settings":           3,
	"Ignore returned URLs?":   18,
	"MQTT Password":           22,
	"MQTT Server":             20,
	"MQTT User":               21,
	"Manual Server Entry":     17,
	"Override Home Assistant and use server chosen (above) for API access.": 19,
	"Please restart the agent to use changed preferences.":                  10,
	"Quit": 4,
	"Save": 9,
	"Select this option to enter a server manually below.": 16,
	"Sensors": 1,
	"The long-lived access token generated in Home Assistant.":                                                                                     12,
	"These are the Home Assistant servers that were detected on the local network.":                                                                14,
	"To register the agent, please enter the relevant details for your Home Assistant\nserver (if not auto-detected) and long-lived access token.": 6,
	"Token":              11,
	"Use Custom Server?": 15,
	"Use MQTT?":          23,
}

var deIndex = []uint32{ // 25 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000,
} // Size: 124 bytes

const deData string = ""

var enIndex = []uint32{ // 25 elements
	0x00000000, 0x00000006, 0x0000000e, 0x0000001b,
	0x00000029, 0x0000002e, 0x0000003f, 0x000000cb,
	0x000000dc, 0x000000ec, 0x000000f1, 0x00000126,
	0x0000012c, 0x00000165, 0x0000017d, 0x000001cb,
	0x000001de, 0x00000213, 0x00000227, 0x0000023d,
	0x00000283, 0x0000028f, 0x00000299, 0x000002a7,
	0x000002b1,
} // Size: 124 bytes

const enData string = "" + // Size: 689 bytes
	"\x02About\x02Sensors\x02App Settings\x02Fyne Settings\x02Quit\x02App Reg" +
	"istration\x02To register the agent, please enter the relevant details fo" +
	"r your Home Assistant\x0aserver (if not auto-detected) and long-lived ac" +
	"cess token.\x02Fyne Preferences\x02App Preferences\x02Save\x02Please res" +
	"tart the agent to use changed preferences.\x02Token\x02The long-lived ac" +
	"cess token generated in Home Assistant.\x02Auto-discovered Servers\x02Th" +
	"ese are the Home Assistant servers that were detected on the local netwo" +
	"rk.\x02Use Custom Server?\x02Select this option to enter a server manual" +
	"ly below.\x02Manual Server Entry\x02Ignore returned URLs?\x02Override Ho" +
	"me Assistant and use server chosen (above) for API access.\x02MQTT Serve" +
	"r\x02MQTT User\x02MQTT Password\x02Use MQTT?"

var frIndex = []uint32{ // 25 elements
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000, 0x00000000, 0x00000000, 0x00000000,
	0x00000000,
} // Size: 124 bytes

const frData string = ""

// Total table size 1061 bytes (1KiB); checksum: B9C99F7E

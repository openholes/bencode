# bencode

![License](https://img.shields.io/badge/license-Apache2.0-green)
![Language](https://img.shields.io/badge/Language-Go-blue.svg)
[![version](https://img.shields.io/github/v/tag/openholes/bencode?label=release&color=blue)](https://github.com/openholes/bencode/releases)
[![Go report](https://goreportcard.com/badge/github.com/openholes/bencode)](https://goreportcard.com/report/github.com/openholes/bencode)
[![Go Reference](https://pkg.go.dev/badge/github.com/openholes/bencode.svg)](https://pkg.go.dev/github.com/openholes/bencode)

`bencode` (pronounced like Bee-encode) is the encoding used by the peer-to-peer file sharing system BitTorrent for storing and transmitting loosely structured data.

## About openHoles

openHoles is an open source organization focusing on peer-to-peer solutions, find more information [here](https://github.com/openholes/openholes)

## Install

```bash
go get github.com/openholes/bencode
```

## Usage

`bencode.Marshal` take a object and returns a byte slice.
`bencode.Unmarshal` take a byte slice and bind data pointer.

```bash
import "github.com/openholes/bencode"

func main() {
	type FileInfo struct {
		Name       string  `bencode:"name"`
		Size       int     `bencode:"size"`
		FloatValue float64 // will be ignored, bencode not support float64 datatype
	}

	type Torrent struct {
		Announce string     `bencode:"announce"`
		Files    []FileInfo `bencode:"files"`
		Created  int64      `bencode:"-"`
	}

	tor := Torrent{
		Announce: "http://tracker",
		Files: []FileInfo{
			{Name: "file1.txt", Size: 1024, FloatValue: 1.0},
		},
	}
	data, err := Marshal(tor)
	if err != nil {
	  // do something
	}
	fmt.Println(string(data)) // d8:announce14:http://tracker5:filesld4:name9:file1.txt4:sizei1024eeee

	var decoded Torrent
	err = Unmarshal(data, &decoded)
	if err != nil {
	  // do something
	}
	fmt.Printf("announce: %s\n", tor.Announce) // announce: http://tracker
}
```

## License

openHoles is licensed under the Apache License 2.0. Refer to [LICENSE](https://github.com/openholes/bencode/blob/main/LICENSE) for more details.

package main

import (
	// Uncomment this line to pass the first stage
	// "encoding/json"
	"encoding/json"
	"fmt"
	"io"
	"os"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

func decode(b string, st int) (x interface{}, i int, err error) {
	if st == len(b) {
		return nil, st, io.ErrUnexpectedEOF
	}
	i = st
	switch {
	case b[i] == 'l':
		return decodeList(b, i)
	case b[i] == 'i':
		return decodeInt(b, i)
	case b[i] >= '0' && b[i] <= '9':
		return decodeString(b, i)
	default:
		return nil, st, fmt.Errorf("unexpected value: %q", b[i])
	}
}

func decodeInt(b string, st int) (x int, i int, err error) {
	i = st
	i++ // 'i'
	if i == len(b) {
		return 0, st, fmt.Errorf("bad int")
	}
	neg := false
	if b[i] == '-' {
		neg = true
		i++
	}
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		x = x*10 + (int(b[i]) - '0')
		i++
	}

	if i == len(b) || b[i] != 'e' {
		return 0, st, fmt.Errorf("bad int")
	}

	i++
	if neg {
		x = -x
	}
	return x, i, nil
}

func decodeString(b string, st int) (x string, i int, err error) {
	var l int

	i = st
	for i < len(b) && b[i] >= '0' && b[i] <= '9' {
		l = l*10 + (int(b[i]) - '0')
		i++
	}
	if i == len(b) || b[i] != ':' {
		return "", st, fmt.Errorf("bad string")
	}
	i++
	if i+l > len(b) {
		return "", st, fmt.Errorf("bad string: out of bounds")
	}
	x = b[i : i+l]
	//	i += l
	i += l
	return x, i, nil
}

func decodeList(b string, st int) (l []interface{}, i int, err error) {
	i = st
	i++ // 'l'
	l = make([]interface{}, 0)
	for {
		if i >= len(b) {
			return nil, st, fmt.Errorf("bad list")
		}
		if b[i] == 'e' {
			i++
			break
		}
		var x interface{}
		x, i, err = decode(b, i)
		if err != nil {
			return nil, i, err
		}
		l = append(l, x)
	}
	return l, i, nil
}

func main() {
	command := os.Args[1]
	if command == "decode" {

		x, _, err := decode(os.Args[2], 0)

		if err != nil {
			fmt.Printf("error: %v\n", err)
			os.Exit(1)
		}
		y, err := json.Marshal(x)
		if err != nil {
			fmt.Printf("error: encode to json%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", y)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

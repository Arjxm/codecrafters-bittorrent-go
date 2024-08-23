package main

import (
	// Uncomment this line to pass the first stage
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

//lli809e5:appleee

func decode(bencode string, start int) (interface{}, error) {
	switch {
	case bencode[start] == 'i':
		return decodeInt(bencode, start)
	case unicode.IsDigit(rune(bencode[start])):
		return decodeString(bencode, start)
	case bencode[start] == 'l':
		return decodeList(bencode, start)
	}
	return nil, fmt.Errorf("invalid bencode")
}

func decodeInt(b string, start int) (interface{}, error) {
	var number string
	for i := start + 1; i < len(b); i++ {
		if b[i] == 'e' {
			break
		}
		number += string(b[i])
	}
	return strconv.Atoi(number)
}

func decodeString(b string, start int) (string, error) {
	var firstColonIndex int
	for i := start; i < len(b); i++ {
		if b[i] == ':' {
			firstColonIndex = i
			break
		}
	}
	lengthStr := b[start:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}
	return b[firstColonIndex+1 : firstColonIndex+1+length], nil
}

func decodeList(b string, start int) ([]interface{}, error) {
	i := start + 1
	decodeList := make([]interface{}, 0)
	for i < len(b) {
		if b[i] == 'e' {
			break
		}
		x, err := decode(b, i)
		if err != nil {
			return nil, err
		}
		decodeList = append(decodeList, x)
		if str, ok := x.(string); ok {
			i += len(str) + len(strconv.Itoa(len(str))) + 2
		} else if num, ok := x.(int); ok {
			i += len(strconv.Itoa(num)) + 2
		} else if list, ok := x.([]interface{}); ok {
			i += len(encodeList(list)) + 2
		}
	}
	return decodeList, nil
}

func encodeList(list []interface{}) string {
	var encoded string
	for _, item := range list {
		if str, ok := item.(string); ok {
			encoded += strconv.Itoa(len(str)) + ":" + str
		} else if num, ok := item.(int); ok {
			encoded += "i" + strconv.Itoa(num) + "e"
		} else if subList, ok := item.([]interface{}); ok {
			encoded += "l" + encodeList(subList) + "e"
		}
	}
	return encoded
}

func main() {
	command := os.Args[1]
	if command == "decode" {
		bencodedValue := os.Args[2]
		decoded, err := decode(bencodedValue, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

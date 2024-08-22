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
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		var firstColonIndex int

		for i := 0; i < len(bencodedString); i++ {
			if bencodedString[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := bencodedString[:firstColonIndex]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if bencodedString[0] == 'i' {

		return strconv.Atoi(bencodedString[1 : len(bencodedString)-1])
	} else if bencodedString[0] == 'l' {
		var decodedList []interface{}

		i := 0
		for i < len(bencodedString) || bencodedString[len(bencodedString)-1] != 'e' {
			if bencodedString[i] == ':' {
				lengthStr := fmt.Sprintf("%c", bencodedString[i-1])
				length, _ := strconv.Atoi(lengthStr)
				decodedList = append(decodedList, bencodedString[i+1:i+1+length])
			}

			if bencodedString[i] == 'i' {
				var number string
				for j := i + 1; j < len(bencodedString)-1; j++ {
					if bencodedString[j] == 'e' {
						i = j
						break
					} else {
						number += string(bencodedString[j])
					}
				}

				num, err := strconv.Atoi(number)
				if err != nil {
					return nil, fmt.Errorf("failed to parse integer value: %v", err)
				}
				decodedList = append(decodedList, num)
			}

			i++
		}

		return decodedList, nil

	} else {
		return "", fmt.Errorf("Only strings are supported at the moment")
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		//		 Uncomment this block to pass the first stage

		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
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

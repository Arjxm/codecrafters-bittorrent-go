package main

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/decoder"
	"github.com/jackpal/bencode-go"
	"os"

	"encoding/json"
	"fmt"
)

type MetaInfo struct {
	Name        string `bencode:"name"`
	Pieces      string `bencode:"pieces"`
	Length      int64  `bencode:"length"`
	PieceLength int64  `bencode:"piece length"`
}
type Meta struct {
	Announce string   `bencode:"announce"`
	Info     MetaInfo `bencode:"info"`
}

func main() {
	command := os.Args[1]

	switch command {

	case "decode":
		x, _, err := decoder.Decode(os.Args[2], 0)

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

	case "info":
		fileName := os.Args[2]
		f, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		var meta Meta
		if err := bencode.Unmarshal(f, &meta); err != nil {
			panic(err)
		}
		fmt.Println("Tracker URL:", meta.Announce)
		fmt.Println("Length:", meta.Info.Length)
		h := sha1.New()
		if err := bencode.Marshal(h, meta.Info); err != nil {
			panic(err)
		}
		fmt.Printf("Info Hash: %x\n", h.Sum(nil))

		fmt.Printf("Piece Length: %d\n", meta.Info.PieceLength)

		fmt.Printf("Piece Hashes: \n")

		for i := 0; i < len(hex.EncodeToString([]byte(meta.Info.Pieces))); {
			j := i + 40
			fmt.Printf("%s\n", hex.EncodeToString([]byte(meta.Info.Pieces))[i:j])
			i = j
		}

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)

	}
}

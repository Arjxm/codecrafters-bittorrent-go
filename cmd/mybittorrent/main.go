package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/bittorrent-starter-go/cmd/decoder"
	"github.com/jackpal/bencode-go"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
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

type TrackerResponse struct {
	Complete    int    `json:"complete"`
	Incomplete  int    `json:"incomplete"`
	Interval    int    `json:"interval"`
	MinInterval int    `json:"min interval"`
	Peers       string `json:"peers"`
}

func getInfo(f *os.File) (Meta, error) {
	var meta Meta
	if err := bencode.Unmarshal(f, &meta); err != nil {
		panic(err)
	}
	return meta, nil
}

func makeRequest(meta Meta) {
	h := sha1.New()
	if err := bencode.Marshal(h, meta.Info); err != nil {
		panic(err)
	}
	infoHash := h.Sum(nil)
	//infoHashBytes, _ := hex.DecodeString(string(infoHash))
	// Query parameters
	params := url.Values{}
	params.Add("info_hash", string(infoHash))
	params.Add("peer_id", "00112233445566778899")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", fmt.Sprint(meta.Info.Length))
	params.Add("compact", "1")

	// Construct the final URL with query parameters
	finalURL := fmt.Sprintf("%s?%s", meta.Announce, params.Encode())

	// Making the GET request
	response, _ := http.Get(finalURL)
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	res, _, _ := decoder.Decode(string(body), 0)

	// Type assert the interface{} to a map[string]interface{}
	data, ok := res.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid response format")
		return
	}

	// Access the "peers" field from the map
	peersValue, ok := data["peers"]
	if !ok {
		fmt.Println("Peers field not found")
		return
	}

	// Type assert the "peers" field to a string
	peers, ok := peersValue.(string)
	if !ok {
		fmt.Println("Invalid peers format")
		return
	}

	for i := 0; i < len(peers); i += 6 {
		ip := net.IP([]byte(peers[i : i+4]))
		port := binary.BigEndian.Uint16([]byte(peers[i+4 : i+6]))

		fmt.Printf("%s:%d\n", ip, port)
	}

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

		meta, _ := getInfo(f)
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

	case "peers":
		fileName := os.Args[2]
		f, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		meta, _ := getInfo(f)
		makeRequest(meta)

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)

	}
}

package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/cmd/decoder"
	"github.com/jackpal/bencode-go"
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

func genrateUrl(meta Meta) string {
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
	return finalURL
}

func getInfo(f *os.File) (Meta, error) {
	var meta Meta
	if err := bencode.Unmarshal(f, &meta); err != nil {
		panic(err)
	}
	return meta, nil
}

func makeRequest(meta Meta) {
	Turl := genrateUrl(meta)
	// Making the GET request
	response, err := http.Get(Turl)
	if err != nil {
		panic(err)
	}

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

	case "handshake":
		fileName := os.Args[2]
		peerAddr := os.Args[3]
		f, err := os.Open(fileName)
		meta, _ := getInfo(f)
		if err != nil {
			panic(err)
		}
		conn, err := net.Dial("tcp", peerAddr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		fmt.Println("Connected to tracker")

		pstrlen := byte(19) // The length of the string "BitTorrent protocol"
		pstr := []byte("BitTorrent protocol")
		reserved := make([]byte, 8) // Eight zeros
		handshake := append([]byte{pstrlen}, pstr...)
		handshake = append(handshake, reserved...)

		h := sha1.New()
		if err := bencode.Marshal(h, meta.Info); err != nil {
			panic(err)
		}
		infoHash := h.Sum(nil)
		handshake = append(handshake, infoHash...)
		handshake = append(handshake, []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}...)

		_, err = conn.Write(handshake)
		buf := make([]byte, 68)
		_, err = conn.Read(buf)
		if err != nil {
			fmt.Println("failed:", err)
			return
		}
		fmt.Printf("Peer ID: %s\n", hex.EncodeToString(buf[48:]))

	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)

	}
}

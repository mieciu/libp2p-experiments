package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
	"log"
	"os"
)

func makeHost(port int) (host.Host, error) {
	privateKey, _, _ := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	hostMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	return libp2p.New(
		libp2p.ListenAddrs(hostMultiAddr),
		libp2p.Identity(privateKey),
	)
}

func startPeer(host host.Host, streamHandler network.StreamHandler) {
	host.SetStreamHandler("/chat/0.0.1", streamHandler)
}

func startPeerAndConnect(host host.Host, targetAddr multiaddr.Multiaddr) (network.Stream, error) {
	targetAddrInfo, _ := peer.AddrInfoFromP2pAddr(targetAddr)
	host.Peerstore().AddAddrs(targetAddrInfo.ID, targetAddrInfo.Addrs, peerstore.PermanentAddrTTL)
	stream, _ := host.NewStream(context.Background(), targetAddrInfo.ID, "/chat/0.0.1")
	return stream, nil
}

func main() {
	sourcePort := flag.Int("port", 0, "Source port number")
	dest := flag.String("dest", "", "Destination multiaddr string")
	flag.Parse()

	host, _ := makeHost(*sourcePort)

	if *dest == "" {
		startPeer(host, handleStream)
		log.Printf("HEY I AM  /ip4/0.0.0.0/tcp/%v/p2p/%s JOIN ME !!!\n", *sourcePort, host.ID().String())
	} else {
		multiAddr, _ := multiaddr.NewMultiaddr(*dest)
		stream, _ := startPeerAndConnect(host, multiAddr)
		handleStream(stream)
	}
	select {}
}

func handleStream(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go readData(rw)
	go writeData(rw)
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		sendData, _ := stdReader.ReadString('\n')
		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')
		if str == "" {
			return
		}
		if str != "\n" {
			fmt.Printf("(INCOMING) %s> ", str)
		}
	}
}

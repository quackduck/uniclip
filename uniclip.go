package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var detectedOs = runtime.GOOS
var listOfClients = make([]*bufio.Writer, 5)
var localClipboard string

func makeServer() {
	l, err := net.Listen("tcp4", "0.0.0.0:")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	port := strconv.FormatInt(int64(l.Addr().(*net.TCPAddr).Port), 10)
	fmt.Println("Run", "`uniclip", getOutboundIP().String()+":"+port+"`", "to join this clipboard")
	fmt.Println()
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleClient(c)
	}
}

func handleClient(c net.Conn) {
	fmt.Println("Connected to:", c.RemoteAddr())
	w := bufio.NewWriter(c)
	listOfClients = append(listOfClients, w)
	defer c.Close()
	go monitorSentClips(bufio.NewReader(c))
	monitorLocalClip(w)
}

func connectToServer(address string) {
	c, err := net.Dial("tcp4", address)
	defer c.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	go monitorSentClips(bufio.NewReader(c))
	monitorLocalClip(bufio.NewWriter(c))
}

func monitorLocalClip(w *bufio.Writer) {
	for {
		localClipboard = getLocalClip()
		// if localClipboard == "" {
		// 	fmt.Println("PANICCCCCCCCCC AHHHHHHHHHHHH IT IS EMPTY")
		// }
		sendClipboard(w, localClipboard)
		for localClipboard == getLocalClip() {
			time.Sleep(time.Second * 5)
		}
		fmt.Println("detected change in clipboard")
	}
}

func monitorSentClips(r *bufio.Reader) {
	var foreignClipboard string
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if s == "STARTCLIPBOARD\n" {
			for {
				s, err = r.ReadString('\n')
				if err != nil {
					fmt.Println(err)
					return
				}
				if s == "ENDCLIPBOARD\n" {
					foreignClipboard = strings.TrimSuffix(foreignClipboard, "\n")
					break
				}
				foreignClipboard += s
			}
			setLocalClip(foreignClipboard)
			localClipboard = foreignClipboard
			fmt.Println("Copied:" + "\n\"" + foreignClipboard + "\"\n")
			for i, w := range listOfClients {
				if w != nil && i != 0 { //don't send to first client, which is this client
					sendClipboard(w, foreignClipboard)
				}
			}
			foreignClipboard = ""
		}
	}
}

func main() {
	if len(os.Args) == 2 {
		connectToServer(os.Args[1])
	} else if len(os.Args) == 1 {
		fmt.Println("Starting a new clipboard!")
		makeServer()
	} else {
		fmt.Println("Too many arguments.\nTo start a new clipboard, use `uniclip`.\nTo connect to a clipboard, use `uniclip <IP>:<PORT>`")
	}
}

func sendClipboard(w *bufio.Writer, clipboard string) {
	var err error
	clipString := "STARTCLIPBOARD\n" + clipboard + "\nENDCLIPBOARD\n"
	_, err = w.WriteString(clipString)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = w.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getLocalClip() string {
	var out []byte
	var err error
	if detectedOs == "darwin" {
		out, err = exec.Command("pbpaste").Output()
	} else if detectedOs == "windows" {
		out, err = exec.Command("powershell.exe", "-command", "Get-Clipboard").Output()
	}
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(out)
}

func setLocalClip(s string) {
	var copyCmd *exec.Cmd
	if detectedOs == "darwin" {
		copyCmd = exec.Command("pbcopy")
	} else if detectedOs == "windows" {
		copyCmd = exec.Command("powershell.exe", "-command", "Set-Clipboard")
	}
	in, err := copyCmd.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := copyCmd.Start(); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := in.Write([]byte(s)); err != nil {
		fmt.Println(err)
		return
	}
	if err := in.Close(); err != nil {
		fmt.Println(err)
		return
	}
	if err := copyCmd.Wait(); err != nil {
		fmt.Println(err)
		return
	}
	return
}

func getOutboundIP() net.IP {
	// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208
	conn, err := net.Dial("udp", "8.8.8.8:80") // address can be anything. Doesn't even have to exist
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var secondsBetweenChecksForClipChange = 1
var helpMsg = `Uniclip - Universal Clipboard
With Uniclip, you can copy from one device and paste on the other.

Usage: uniclip [ <address> | --help/-h ]
Examples:
   uniclip                          # start a new clipboard
   uniclip 192.168.86.24:53701      # join the clipboard at the address - 192.168.86.24:53701
   uniclip --help                   # print this help message

Running just ` + "`uniclip`" + ` will start a new clipboard.
It will also provide an address with which you can connect to the clipboard with another device.`

var detectedOs = runtime.GOOS
var listOfClients = make([]*bufio.Writer, 0)
var localClipboard string
var lock sync.Mutex

func main() {
	if len(os.Args) == 2 {
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Println(helpMsg)
			return
		}
		connectToServer(os.Args[1])
	} else if len(os.Args) == 1 {
		fmt.Println("Starting a new clipboard!")
		makeServer()
	} else {
		fmt.Println("Too many arguments.\nTo start a new clipboard, use `uniclip`.\nTo connect to a clipboard, use `uniclip <IP>:<PORT>`")
	}
}

func makeServer() {
	l, err := net.Listen("tcp4", "0.0.0.0:")
	if err != nil {
		handleError(err)
		return
	}
	defer l.Close()
	port := strconv.FormatInt(int64(l.Addr().(*net.TCPAddr).Port), 10)
	fmt.Println("Run", "`uniclip", getOutboundIP().String()+":"+port+"`", "to join this clipboard")
	fmt.Println()
	for {
		c, err := l.Accept()
		if err != nil {
			handleError(err)
			return
		}
		fmt.Println("Connected to a client")
		go handleClient(c)
	}
}

func handleClient(c net.Conn) {
	w := bufio.NewWriter(c)
	listOfClients = append(listOfClients, w)
	defer c.Close()
	monitorSentAndLocal(bufio.NewReader(c), w)
}

func connectToServer(address string) {
	c, err := net.Dial("tcp4", address)
	defer c.Close()
	if err != nil {
		handleError(err)
		return
	}
	fmt.Println("Connected to the clipboard")
	monitorSentAndLocal(bufio.NewReader(c), bufio.NewWriter(c))
}

// func monitorLocalClip(w *bufio.Writer) {
// 	for {
// 		lock.Lock()
// 		localClipboard = getLocalClip()
// 		lock.Unlock()
// 		sendClipboard(w, localClipboard)
// 		for localClipboard == getLocalClip() {
// 			time.Sleep(time.Second * time.Duration(secondsBetweenChecksForClipChange))
// 		}
// 	}
// }

func monitorSentAndLocal(r *bufio.Reader, w *bufio.Writer) {
	var foreignClipboard string
	for {
		localClipboard = getLocalClip()
		err := sendClipboard(w, localClipboard)
		if err != nil {
			handleError(err)
			return
		}
		time.Sleep(time.Second * time.Duration(secondsBetweenChecksForClipChange))
		s, err := r.ReadString('\n')
		if err != nil {
			handleError(err)
			return
		}
		if s == "STARTCLIPBOARD\n" {
			for {
				s, err = r.ReadString('\n')
				if err != nil {
					handleError(err)
					return
				}
				if s == "ENDCLIPBOARD\n" {
					foreignClipboard = strings.TrimSuffix(foreignClipboard, "\n")
					break
				}
				foreignClipboard += s
			}
			setLocalClip(foreignClipboard)
			// localClipboard = foreignClipboard
			fmt.Println("Copied:" + "\n\"" + foreignClipboard + "\"\n")
			fmt.Println(listOfClients)
			for i, w := range listOfClients {
				if w != nil { // && i != 0 { // don't send to first client, which is this client
					err := sendClipboard(w, foreignClipboard)
					if err != nil {
						listOfClients[i] = nil
						handleError(err)
					}
				}
			}
			foreignClipboard = ""
		}
	}
}

func sendClipboard(w *bufio.Writer, clipboard string) error {
	var err error
	clipString := "STARTCLIPBOARD\n" + clipboard + "\nENDCLIPBOARD\n"
	_, err = w.WriteString(clipString)
	if err != nil {
		return err
	}
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func getLocalClip() string {
	var out []byte
	var err error
	var cmd *exec.Cmd
	errMsg := "An error occurred wile getting the local clipboard"
	if detectedOs == "darwin" {
		cmd = exec.Command("pbpaste")
	} else if detectedOs == "windows" {
		cmd = exec.Command("powershell.exe", "-command", "Get-Clipboard")
	}
	if out, err = cmd.CombinedOutput(); err != nil {
		handleError(err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(exiterr.Stderr))
		}
		return errMsg
	}
	if detectedOs == "windows" {
		return strings.TrimSuffix(string(out), "\n") // ps's get-clipboard adds a newline to the end for some reason
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
		handleError(err)
		return
	}
	if err := copyCmd.Start(); err != nil {
		handleError(err)
		return
	}
	if _, err := in.Write([]byte(s)); err != nil {
		handleError(err)
		return
	}
	if err := in.Close(); err != nil {
		handleError(err)
		return
	}
	if err := copyCmd.Wait(); err != nil {
		handleError(err)
		return
	}
	return
}

func getOutboundIP() net.IP {
	// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go/37382208#37382208
	conn, err := net.Dial("udp", "8.8.8.8:80") // address can be anything. Doesn't even have to exist
	if err != nil {
		handleError(err)
		return nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func handleError(err error) {
	if err == io.EOF {
		fmt.Println("Disconnected from a device")
	} else {
		fmt.Println("An error occurred:", err)
	}
	return
}

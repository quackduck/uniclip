package main

import (
	"bufio"
	"errors"
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
With Uniclip, you can copy from one device and paste on another.

Usage: uniclip [--verbose/-v] [ <address> | --help/-h ]
Examples:
   uniclip                          # start a new clipboard
   uniclip 192.168.86.24:53701      # join the clipboard at the address - 192.168.86.24:53701
   uniclip --help                   # print this help message
   uniclip -v 192.168.86.24:53701   # enable verbose output 
Running just ` + "`uniclip`" + ` will start a new clipboard.
It will also provide an address with which you can connect to the same clipboard with another device.
Refer to https://github.com/quackduck/uniclip for more information`

var detectedOs = runtime.GOOS
var listOfClients = make([]*bufio.Writer, 0)
var localClipboard string
var lock sync.Mutex
var debug = false

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	if len(os.Args) == 2 || len(os.Args) == 3 {
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			fmt.Println(helpMsg)
			return
		}
		if os.Args[1] == "--verbose" || os.Args[1] == "-v" {
			debug = true
			os.Args = os.Args[1:]
			main()
			return
		}
		connectToServer(os.Args[1])
	} else if len(os.Args) == 1 {
		makeServer()
	} else {
		handleError(errors.New("too many arguments"))
		fmt.Println(helpMsg)
	}
}

func makeServer() {
	fmt.Println("Starting a new clipboard")
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
		fmt.Println("Connected to a device")
		go handleClient(c)
	}
}

func handleClient(c net.Conn) {
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
		handleError(err)
		return
	}
	fmt.Println("Connected to the clipboard")
	go monitorSentClips(bufio.NewReader(c))
	monitorLocalClip(bufio.NewWriter(c))
}

func monitorLocalClip(w *bufio.Writer) {
	for {
		lock.Lock()
		localClipboard = getLocalClip()
		lock.Unlock()
		if debug {
			fmt.Println("clipboard changed; sending new one. localClipboard =", localClipboard)
		}
		err := sendClipboard(w, localClipboard)
		if err != nil {
			handleError(err)
			return
		}
		for localClipboard == getLocalClip() {
			time.Sleep(time.Second * time.Duration(secondsBetweenChecksForClipChange))
		}
	}
}

func monitorSentClips(r *bufio.Reader) {
	var foreignClipboard string
	for {
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
			lock.Lock() // don't let the goroutine that checks for changes do anything
			setLocalClip(foreignClipboard)
			localClipboard = foreignClipboard
			lock.Unlock() // we've made sure the other goroutine won't have a false positive
			if debug {
				fmt.Println("rcvd:", foreignClipboard) // debugging
				if foreignClipboard == "" {
					fmt.Println("received empty string")
				}
			}
			for i, w := range listOfClients {
				if w != nil {
					err := sendClipboard(w, foreignClipboard)
					if err != nil {
						listOfClients[i] = nil
						fmt.Println("error when trying to send the clipboard to a device. Will not contact that device again.")
						//handleError(err)
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
	if debug {
		if clipboard == "" {
			fmt.Println("was going to send empty string but skipping")
			return nil
		}
		fmt.Println("sent:", clipboard) //debugging
	}
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
	} else {
		// Unix - check what's available
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-out", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--output", "--clipboard")
		} else if _, err := exec.LookPath("wl-paste"); err == nil {
			cmd = exec.Command("wl-paste", "--no-newline")
		} else if _, err := exec.LookPath("termux-clipboard-get"); err == nil {
			cmd = exec.Command("termux-clipboard-get")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	if out, err = cmd.CombinedOutput(); err != nil {
		handleError(err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(exiterr.Stderr))
		}
		return errMsg
	}
	if detectedOs == "windows" {
		return strings.TrimSuffix(string(out), "\r\n") // powershell's get-clipboard adds a windows newline to the end for some reason
	}
	return string(out)
}

func setLocalClip(s string) {
	var copyCmd *exec.Cmd
	if detectedOs == "darwin" {
		copyCmd = exec.Command("pbcopy")
	} else if detectedOs == "windows" {
		copyCmd = exec.Command("powershell.exe", "-command", "Set-Clipboard -Value "+"\""+s+"\"")
	} else {
		if _, err := exec.LookPath("xclip"); err == nil {
			copyCmd = exec.Command("xclip", "-in", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			copyCmd = exec.Command("xsel", "--input", "--clipboard")
		} else if _, err := exec.LookPath("wl-paste"); err == nil {
			copyCmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("termux-clipboard-set"); err == nil {
			copyCmd = exec.Command("termux-clipboard-set")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux:API installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	var in io.WriteCloser
	var err error
	if detectedOs != "windows" { // the clipboard has already been set in the arguments for the windows command
		in, err = copyCmd.StdinPipe()
		if err != nil {
			handleError(err)
			return
		}
	}
	if err = copyCmd.Start(); err != nil {
		handleError(err)
		return
	}
	if detectedOs != "windows" {
		if _, err = in.Write([]byte(s)); err != nil {
			handleError(err)
			return
		}
		if err = in.Close(); err != nil {
			handleError(err)
			return
		}
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
		_, _ = fmt.Fprintln(os.Stderr, "uniclip: error:", err)
	}
	return
}

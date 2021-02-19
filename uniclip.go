package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/gob"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/scrypt"
)

var (
	secondsBetweenChecksForClipChange = 1
	helpMsg                           = `Uniclip - Universal Clipboard
With Uniclip, you can copy from one device and paste on another.

Usage: uniclip [--debug/-d] [ <address> | --help/-h ]
Examples:
   uniclip                          # start a new clipboard
   uniclip 192.168.86.24:53701      # join the clipboard at 192.168.86.24:53701
   uniclip -d                       # start a new clipboard with debug output
   uniclip -d 192.168.86.24:53701   # join the clipboard with debug output
Running just ` + "`uniclip`" + ` will start a new clipboard.
It will also provide an address with which you can connect to the same clipboard with another device.
Refer to https://github.com/quackduck/uniclip for more information`
	listOfClients  = make([]*bufio.Writer, 0)
	localClipboard string
	printDebugInfo = false
	version        = "v2.0.1"
	cryptoStrength = 16384
	secure         = false
	password       []byte
)

// TODO: Add a way to reconnect (if computer goes to sleep)
func main() {
	if len(os.Args) > 4 {
		handleError(errors.New("too many arguments"))
		fmt.Println(helpMsg)
		return
	}
	if hasOption, _ := argsHaveOption("help", "h"); hasOption {
		fmt.Println(helpMsg)
		return
	}
	if hasOption, i := argsHaveOption("debug", "d"); hasOption {
		printDebugInfo = true
		os.Args = removeElemFromSlice(os.Args, i) // delete the debug option and run again
		main()
		return
	}
	// --secure encrypts your data
	if hasOption, i := argsHaveOption("secure", "s"); hasOption {
		secure = true
		os.Args = removeElemFromSlice(os.Args, i) // delete the secure option and run again
		fmt.Print("Password: ")
		password, _ = terminal.ReadPassword(syscall.Stdin)
		fmt.Println()
		main()
		return
	}
	if hasOption, _ := argsHaveOption("version", "v"); hasOption {
		fmt.Println(version)
		return
	}
	if len(os.Args) == 2 { // has exactly one argument
		connectToServer(os.Args[1])
		return
	}
	makeServer()
}

func makeServer() {
	fmt.Println("Starting a new clipboard")
	l, err := net.Listen("tcp4", ":") //nolint // complains about binding to all interfaces
	if err != nil {
		handleError(err)
		return
	}
	defer l.Close()
	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
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
	if c == nil {
		handleError(err)
		fmt.Println("Could not connect to", address)
		return
	}
	if err != nil {
		handleError(err)
		return
	}
	defer func() { _ = c.Close() }()
	fmt.Println("Connected to the clipboard")
	go monitorSentClips(bufio.NewReader(c))
	monitorLocalClip(bufio.NewWriter(c))
}

func monitorLocalClip(w *bufio.Writer) {
	for {
		localClipboard = getLocalClip()
		//debug("clipboard changed so sending it. localClipboard =", localClipboard)
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
	var foreignClipboardBytes []byte
	for {
		err := gob.NewDecoder(r).Decode(&foreignClipboardBytes)
		if err != nil {
			if err == io.EOF {
				return // no need to monitor: disconnected
			}
			handleError(err)
			continue // continue getting next message
		}
		if secure {
			foreignClipboardBytes, err = decrypt(password, foreignClipboardBytes)
		}
		foreignClipboard = string(foreignClipboardBytes)

		setLocalClip(foreignClipboard)
		localClipboard = foreignClipboard
		debug("rcvd:", foreignClipboard)
		for i := range listOfClients {
			if listOfClients[i] != nil {
				err = sendClipboard(listOfClients[i], foreignClipboard)
				if err != nil {
					listOfClients[i] = nil
					fmt.Println("Error when trying to send the clipboard to a device. Will not contact that device again.")
				}
			}
		}
		foreignClipboard = ""
	}
}

func sendClipboard(w *bufio.Writer, clipboard string) error {
	var clipboardBytes []byte
	var err error
	if secure {
		clipboardBytes, err = encrypt(password, []byte(clipboard))
		if err != nil {
			return err
		}
	} else {
		clipboardBytes = []byte(clipboard)
	}
	err = gob.NewEncoder(w).Encode(clipboardBytes)
	if err != nil {
		return err
	}
	debug("sent:", clipboard)
	//if secure {
	//	debug("--secure is enabled, so actually sent as:", hex.EncodeToString(clipboardBytes))
	//}
	return w.Flush()
}

// Thanks to https://bruinsslot.jp/post/golang-crypto/ for crypto logic
func encrypt(key, data []byte) ([]byte, error) {
	key, salt, err := deriveKey(key, nil)
	if err != nil {
		return nil, err
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	ciphertext = append(ciphertext, salt...)
	return ciphertext, nil
}

func decrypt(key, data []byte) ([]byte, error) {
	salt, data := data[len(data)-32:], data[:len(data)-32]
	key, _, err := deriveKey(key, salt)
	if err != nil {
		return nil, err
	}
	blockCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {
		return nil, err
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func deriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}
	key, err := scrypt.Key(password, salt, cryptoStrength, 8, 1, 32)
	if err != nil {
		return nil, nil, err
	}
	return key, salt, nil
}

func getLocalClip() string {
	var out []byte
	var err error
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows": //nolint // complains about literal string "windows" being used multiple times
		cmd = exec.Command("powershell.exe", "-command", "Get-Clipboard")
	default:
		if _, err = exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-out", "-selection", "clipboard")
		} else if _, err = exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--output", "--clipboard")
		} else if _, err = exec.LookPath("wl-paste"); err == nil {
			cmd = exec.Command("wl-paste", "--no-newline")
		} else if _, err = exec.LookPath("termux-clipboard-get"); err == nil {
			cmd = exec.Command("termux-clipboard-get")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	if out, err = cmd.Output(); err != nil {
		handleError(err)
		return "An error occurred wile getting the local clipboard"
	}
	if runtime.GOOS == "windows" {
		return strings.TrimSuffix(string(out), "\r\n") // powershell's get-clipboard adds a windows newline to the end for some reason
	}
	return string(out)
}

func setLocalClip(s string) {
	var copyCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		copyCmd = exec.Command("pbcopy")
	case "windows":
		copyCmd = exec.Command("powershell.exe", "-command", "Set-Clipboard") //-Value "+"\""+s+"\"")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			copyCmd = exec.Command("xclip", "-in", "-selection", "clipboard")
		} else if _, err = exec.LookPath("xsel"); err == nil {
			copyCmd = exec.Command("xsel", "--input", "--clipboard")
		} else if _, err = exec.LookPath("wl-copy"); err == nil {
			copyCmd = exec.Command("wl-copy")
		} else if _, err = exec.LookPath("termux-clipboard-set"); err == nil {
			copyCmd = exec.Command("termux-clipboard-set")
		} else {
			handleError(errors.New("sorry, uniclip won't work if you don't have xsel, xclip, wayland or Termux:API installed :(\nyou can create an issue at https://github.com/quackduck/uniclip/issues"))
			os.Exit(2)
		}
	}
	in, err := copyCmd.StdinPipe()
	if err != nil {
		handleError(err)
		return
	}
	if err = copyCmd.Start(); err != nil {
		handleError(err)
		return
	}
	if runtime.GOOS != "windows" {
		if _, err = in.Write([]byte(s)); err != nil {
			handleError(err)
			return
		}
		if err = in.Close(); err != nil {
			handleError(err)
			return
		}
	}
	if err = copyCmd.Wait(); err != nil {
		handleError(err)
		return
	}
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
		fmt.Println("Disconnected")
	} else {
		fmt.Fprintln(os.Stderr, "error: ["+err.Error()+"]")
	}
}

func debug(a ...interface{}) {
	if printDebugInfo {
		fmt.Println("verbose:", a)
	}
}

func argsHaveOption(long string, short string) (hasOption bool, foundAt int) {
	for i, arg := range os.Args {
		if arg == "--"+long || arg == "-"+short {
			return true, i
		}
	}
	return false, 0
}

// keep order
func removeElemFromSlice(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}

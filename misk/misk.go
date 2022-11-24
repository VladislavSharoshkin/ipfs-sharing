package misk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mr-tron/base58"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^\p{L}\p{N} ]+`)

func structToMap(v interface{}) (map[string]interface{}, error) {
	vMap := &map[string]interface{}{}

	err := mapstructure.Decode(v, &vMap)
	if err != nil {
		return nil, err
	}

	return *vMap, nil
}

func PrintValue(name string, values ...interface{}) {
	//if len(values) > 0 {
	//	name += ":"
	//}

	fmt.Println(strings.ToUpper(name), fmt.Sprint(values...)) // color.CyanString(name)

}

func SPrintValues(values ...interface{}) string {
	text := fmt.Sprintln(values...)
	return text[:len(text)-1]
}

func Sha(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func ShaString(data string) string {
	return hex.EncodeToString(Sha([]byte(data)))
}

func ClearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, " ")
}

func RandomBytes(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

func RandomString() string {
	return base58.Encode(RandomBytes(32))
}

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func OpenFolder(dir string) {
	cmd := "open"
	if runtime.GOOS == "windows" {
		cmd = "explorer"
	}
	err := exec.Command(cmd, dir).Start()
	if err != nil {
		log.Println(err)
	}
}

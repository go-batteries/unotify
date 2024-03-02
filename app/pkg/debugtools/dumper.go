package debugtools

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
)

func Logdeep(v interface{}, args ...string) {
	if len(args) == 0 {
		args = append(args, "error")
	}

	message := strings.Join(args, " ")

	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		log.Printf("%+v\n", v)
		return
	}

	log.Println(message, string(b))
}

func HttpRequestLog(req *http.Request) {
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Printf("failed to dump %v\n", err)
		return
	}

	log.Println(string(b))
}

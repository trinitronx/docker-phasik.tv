// HTTP Server for phasik.tv
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// Startup message returned to the console when the server starts
const startupMessage = `` +
	`                 ___________________` + "\n" +
	`                /   ________ -[]--. \` + "\n" +
	"               / ,-'         `-.   \\ \\" + "\n" +
	`              / (       o       )  _) \` + "\n" +
	"             /   `-._________,-'_ /_/-.\\" + "\n" +
	`            /  __ _   Phasik   " " "    \` + "\n" +
	`           /_____________________________\` + "\n" +
	`           "-=-------------------------=-"` + "\n" +
	`		   phasik.tv started!`

type httpMinimalResponse struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Data       []byte `json:"data"`
}

func serveFiles(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./" {
		p = "./static/index.html"
	}
	http.ServeFile(w, r, p)
}

// response2JSON is a helper function to convert HTTP status codes to JSON
// JSON output is written to the provided io.PipeWriter
func response2JSON(status uint16, in_r *io.PipeReader, out_w *io.PipeWriter) {
	// func response2JSON(status uint16, in_r *io.PipeReader, out_w *http.ResponseWriter) {
	response := httpMinimalResponse{
		Status:     http.StatusText(int(status)),
		StatusCode: int(status),
		Data:       []byte(""),
	}

	json_enc := json.NewEncoder(out_w)
	// json_enc := json.NewEncoder(*out_w)
	err := json.NewDecoder(in_r).Decode(&(&response).Data)
	if err != nil && err != io.EOF {
		fmt.Printf("response2JSON error: %+v\n", err)
		// out_w.CloseWithError(err)
		json_enc.Encode(httpMinimalResponse{Status: http.StatusText(http.StatusInternalServerError), StatusCode: int(http.StatusInternalServerError), Data: []byte(err.Error())})
	}

	// ok := json_enc.Encode(map[string]string{"status": "ok"})
	json_enc.Encode(&response)
	out_w.Close()
	// *out_w.Close()
	// out_w.Write()
	// http.ResponseWriter.Write(*out_w, []byte{0})
	// http.ResponseWriter.Write(*out_w, response.Data)
	//Write([]byte("test"))
	fmt.Printf("response2JSON: %+v\n", response)
}

// main is the entrypoint for the phasik.tv server
func main() {

	// logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)

	// const json_ok, var _err = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// json_ok :=  // json.NewEncoder(json_pipe).Encode(map[string]string{"status": "ok"})
	// json.NewEncoder()
	fs := http.FileServer(http.Dir("/srv/www"))
	http.Handle("/", fs)

	http.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s %s\n", r.Method, r.URL.Path)
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
	})
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		json_pipe_r, json_pipe_w := io.Pipe()
		data_pipe_r, data_pipe_w := io.Pipe()
		// fmt.Fprintf(w, "%s %s\n", r.Method, r.URL.Path)
		// w.WriteHeader(http.StatusOK)
		// { status: "ok" }
		// json_ok := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		// fmt.Fprintf(w, json_ok)
		data_pipe_w.Close()
		go response2JSON(http.StatusOK, data_pipe_r, json_pipe_w)
		// go response2JSON(http.StatusOK, data_pipe_r, &w)

		fmt.Printf("%s %s -> ", r.Method, r.URL.Path)

		// var json_resp []byte
		// _, err := json_pipe_r.Read(json_resp) // Copilot said: json_pipe_r.Read(json_resp) is not the correct way to read the response from the pipe.
		json_resp, err := io.ReadAll(json_pipe_r)
		// if err != nil && err != io.EOF {
		if err != nil {
			fmt.Println(err)
		}

		var resp httpMinimalResponse
		json.Unmarshal(json_resp, &resp)

		fmt.Printf("resp (unmarshalled): %+v\n", resp)
		fmt.Printf("json_resp: %s\n", json_resp)
		// logger.Printf("json_resp: %s", json_resp)

		// send JSON response
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(json_resp))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	config_file := os.Getenv("CONFIG_FILE")
	if config_file == "" {
		config_file = "config.yml"
	}

	for _, line := range strings.Split(startupMessage, "\n") {
		fmt.Println(line)
	}
	fmt.Printf("Server listening at :%s 🚀\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}

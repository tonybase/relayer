package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

func getAPICmd() *cobra.Command {
	apiCmd := &cobra.Command{
		Use: "api",
		// Aliases: []string{""},
		Short: "Start the relayer API",
		RunE: func(cmd *cobra.Command, args []string) error {
			http.HandleFunc("/", handleExec)
			log.Println("listening on", config.Global.APIListenAddr)
			return http.ListenAndServe(config.Global.APIListenAddr, nil)
		},
	}
	return apiCmd
}

var cmdPermits = []string{"tx"}

type errorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Output  string `json:"output"`
}

func handleExec(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		hasPermit bool
		outWriter = new(bytes.Buffer)
		args      = strings.Split(strings.TrimLeft(r.URL.Path, "/"), "/")
	)
	if len(args) <= 1 {
		handleWrite(w, r, errorReply{Code: http.StatusBadRequest, Message: fmt.Sprintf("invalid args:%s", args)})
		return
	}
	for _, w := range cmdPermits {
		if w == args[0] {
			hasPermit = true
		}
	}
	if !hasPermit {
		handleWrite(w, r, errorReply{Code: http.StatusUnauthorized, Message: fmt.Sprintf("unauthorized args:%s", args)})
		return
	}
	rootCmd.SetArgs(args)
	rootCmd.SetOutput(outWriter)
	if err = rootCmd.ExecuteContext(r.Context()); err != nil {
		handleWrite(w, r, errorReply{Code: http.StatusInternalServerError, Message: err.Error()})
	} else {
		output := outWriter.String()
		if strings.HasPrefix(output, "Commands") {
			handleWrite(w, r, errorReply{Code: http.StatusBadRequest, Message: fmt.Sprintf("invalid args:%s", args), Output: output})
		} else {
			handleWrite(w, r, errorReply{Code: http.StatusOK, Message: "OK", Output: output})
		}
	}
}

func handleWrite(w http.ResponseWriter, r *http.Request, reply errorReply) {
	b, _ := json.Marshal(reply)
	w.WriteHeader(reply.Code)
	w.Write(b)
	// access logs
	log.Printf("handleWrite path: %s returns: %s", r.URL.Path, b)
}
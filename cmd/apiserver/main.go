package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"

	emailVerifier "github.com/AfterShip/email-verifier"
)

func GetEmailVerification(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	verifier := emailVerifier.NewVerifier()
	ret, err := verifier.Verify(ps.ByName("email"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !ret.Syntax.Valid {
		_, _ = fmt.Fprint(w, "email address syntax is invalid")
		return
	}

	bytes, err := json.Marshal(ret)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, _ = fmt.Fprint(w, string(bytes))

}

func main() {
	router := httprouter.New()

	router.GET("/v1/:email/verification", GetEmailVerification)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Println("Listening on port 8080...")

	log.Fatal(server.ListenAndServe())
}

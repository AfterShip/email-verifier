package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	emailVerifier "github.com/AfterShip/email-verifier"
	"github.com/julienschmidt/httprouter"
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
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

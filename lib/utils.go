package lib

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
)

func GenerateHMAC(msg []byte, key string) string {
	mac := hmac.New(sha512.New, []byte(key))
	mac.Write(msg)
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidateSignature(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(400), 400)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		key := os.Getenv("TFC_RUN_TASK_HMAC_KEY")
		origSig := r.Header.Get("x-tfc-task-signature")
		calcSig := GenerateHMAC(body, key)
		if origSig != calcSig {
			log.Printf("[ERROR] signature mismatch got: %s,expected: %s\n", origSig, calcSig)
			http.Error(w, http.StatusText(400), 400)
			return
		}
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Hello World"))
}

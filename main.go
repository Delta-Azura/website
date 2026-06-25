// main.go
package main

import (
    "database/sql"
    "html/template"
    "net/http"
    _ "github.com/lib/pq"
)

var db *sql.DB
var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func main() {
    var err error
    db, err = sql.Open("postgres", "user=coiffeuse password=motdepasse dbname=rdv sslmode=disable")
    if err != nil {
        panic(err)
    }
    err = db.Ping()
    if err != nil {
        panic(err)
    }
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/creneaux", creneauxHandler)
    http.HandleFunc("/reserver", reserverHandler)
    http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "index.html", nil)
}

func creneauxHandler(w http.ResponseWriter, r *http.Request) {
    rows, _ := db.Query(
        "SELECT id, heure FROM creneaux WHERE disponible = true ORDER BY heure",
    )
    defer rows.Close()

    type Creneau struct{ ID int; Heure string }
    var creneaux []Creneau
    for rows.Next() {
        var c Creneau
        rows.Scan(&c.ID, &c.Heure)
        creneaux = append(creneaux, c)
    }
    tmpl.ExecuteTemplate(w, "creneaux.html", creneaux)
}

func reserverHandler(w http.ResponseWriter, r *http.Request) {
    id := r.FormValue("id")
    nom := r.FormValue("nom")
    db.Exec(
        "UPDATE creneaux SET disponible=false, client=$1 WHERE id=$2", nom, id,
    )
    w.Write([]byte(`<p class="success">Réservation confirmée !</p>`))
}
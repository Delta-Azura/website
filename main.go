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
    date := r.FormValue("date")
    if date == "" {
        w.Write([]byte(`<p class="error">Veuillez sélectionner une date.</p>`))
        return
    }
    touslescreneaux := []string{"09:00", "10:30", "14:00", "16:00"}
    rows, err := db.Query(
        "SELECT heure FROM creneaux WHERE disponible = false AND date = $1",
        date,
    )
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    reserved := map[string]bool{}
    for rows.Next() {
        var heure string
        rows.Scan(&heure)
        reserved[heure] = true
    }
    type Creneau struct{ Heure string; Date string }
    var creneaux []Creneau
    for _, h := range touslescreneaux {
        if !reserved[h] {
            creneaux = append(creneaux, Creneau{Heure: h, Date: date})
        }

    }
    tmpl.ExecuteTemplate(w, "creneaux.html", creneaux)
}

func reserverHandler(w http.ResponseWriter, r *http.Request) {
    heure := r.FormValue("heure")
    nom := r.FormValue("nom")
    date := r.FormValue("date")

    db.Exec(
        "INSERT INTO creneaux (heure, date, disponible, client) VALUES ($1, $2, false, $3)",
        heure, date, nom,
    
    )
    w.Write([]byte(`<p class="success">Réservation confirmée !</p>`))
}
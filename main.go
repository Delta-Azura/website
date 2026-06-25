// main.go
package main

import (
    "database/sql"
    "html/template"
    "net/http"
    "time"
    "os"
    "github.com/gorilla/sessions"
    _ "github.com/lib/pq"
)

var db *sql.DB
var tmpl = template.Must(template.ParseGlob("templates/*.html"))
var store = sessions.NewCookieStore([]byte("secret-key"))
var adminUser = os.Getenv("ADMIN_USER")
var adminPass = os.Getenv("ADMIN_PASS")



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
    http.Handle("/static/", http.FileServer(http.Dir(".")))
    http.HandleFunc("/admin", adminHandler) 
    http.HandleFunc("/admin/login", adminLoginHandler)
    http.HandleFunc("/admin/logout", adminLogoutHandler)
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/creneaux", creneauxHandler)
    http.HandleFunc("/reserver", reserverHandler)
    http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    tmpl.ExecuteTemplate(w, "index.html", nil)
}


func adminLoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        user := r.FormValue("username")
        pass := r.FormValue("password")
        if user == adminUser && pass == adminPass {
            session, _ := store.Get(r, "session")
            session.Values["auth"] = true
            session.Save(r, w)
            http.Redirect(w, r, "/admin", http.StatusSeeOther)
            return
        }
        tmpl.ExecuteTemplate(w, "login.html", "Nom d'utilisateur ou mot de passe incorrect.")
    } else {
        tmpl.ExecuteTemplate(w, "login.html", nil)
    }
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    if auth, ok := session.Values["auth"].(bool); !ok || !auth {
        http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
        return
    }
    date := r.URL.Query().Get("date")
    if date == "" {
        date = time.Now().Format("2006-01-02")
    }
    rows, err := db.Query("SELECT heure, client FROM creneaux WHERE date = $1 AND disponible = false", date)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    type RDV struct{ Heure string; Client string }
    var rdvs []RDV
    for rows.Next() {
        var rdv RDV
        rows.Scan(&rdv.Heure, &rdv.Client)
        rdvs = append(rdvs, rdv)
    }
    tmpl.ExecuteTemplate(w, "admin.html", map[string]interface{}{
        "RDVs": rdvs,
        "Date": date,
    })
}


func adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session")
    session.Values["auth"] = false
    session.Save(r, w)
    http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
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
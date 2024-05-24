package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"strconv"
)

type RequestBody struct {
	Message string `json:"message"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Vinyl struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Releasedate int32   `json:"releasedate"`
	Price       int32   `json:"price"`
	Rating      float32 `json:"rating"`
}

func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	vinyls, err := GetVinylsFromDB()
	if err != nil {
		http.Error(w, "Failed to fetch vinyls", http.StatusInternalServerError)
		return
	}
	filter := r.URL.Query().Get("filter")
	sort := r.URL.Query().Get("sort")
	page := r.URL.Query().Get("page")

	limit := 10
	offset := 0

	if p, err := strconv.Atoi(page); err == nil && p > 1 {
		offset = (p - 1) * limit
	}

	log.WithFields(logrus.Fields{
		"action": "mainPageHandler",
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Handling main page request")

	query := "SELECT id, title, artist, releasedate, price, rating FROM vinyls"
	if filter != "" {
		query += " WHERE brand LIKE '%" + filter + "%'"
	}
	if sort != "" {
		query += " ORDER BY " + sort
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	vinyls, err = GetVinylsFromDBWithPagination(query)
	if err != nil {
		http.Error(w, "Failed to fetch vinyls", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("public/index.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, vinyls)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

}

func GetVinylsFromDBWithPagination(query string) ([]Vinyl, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName)
	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vinyls []Vinyl
	for rows.Next() {
		var vinyl Vinyl
		if err := rows.Scan(&vinyl.ID, &vinyl.Title, &vinyl.Artist, &vinyl.Releasedate, &vinyl.Price, &vinyl.Rating); err != nil {
			return nil, err
		}
		vinyls = append(vinyls, vinyl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vinyls, nil
}

func GetVinylsFromDB() ([]Vinyl, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(localhost:3306)/%s", dbUser, dbPass, dbName)
	db, err := sql.Open(dbDriver, dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, artist, releasedate, price, rating FROM vinyls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vinyls []Vinyl
	for rows.Next() {
		var vinyl Vinyl
		if err := rows.Scan(&vinyl.ID, &vinyl.Title, &vinyl.Artist, &vinyl.Releasedate, &vinyl.Price, &vinyl.Rating); err != nil {
			return nil, err
		}
		vinyls = append(vinyls, vinyl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return vinyls, nil
}

func createVinylHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	releasedate, err := strconv.Atoi(r.FormValue("releasedate"))
	price, err := strconv.Atoi(r.FormValue("price"))
	rating, err := strconv.ParseFloat(r.FormValue("rating"), 32)

	err = CreateVinyl(db, title, artist, int32(releasedate), int32(price), float32(rating))
	if err != nil {
		http.Error(w, "Failed to create vinyl", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func CreateVinyl(db *sql.DB, title, artist string, releasedate, price int32, rating float32) error {
	query := "INSERT INTO vinyls (title, artist, releasedate, price, rating) VALUES (?, ?, ?,?,?)"
	_, err := db.Exec(query, title, artist, releasedate, price, rating)
	if err != nil {
		return err
	}
	return nil
}

func getVinylHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	vars := mux.Vars(r)
	idStr := vars["id"]

	vinylID, err := strconv.Atoi(idStr)

	user, err := GetVinyl(db, vinylID)
	if err != nil {
		http.Error(w, "Vinyl not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func GetVinyl(db *sql.DB, id int) (*Vinyl, error) {
	query := "SELECT * FROM vinyls WHERE id = ?"
	row := db.QueryRow(query, id)

	vinyl := &Vinyl{}
	err := row.Scan(&vinyl.ID, &vinyl.Title, &vinyl.Artist, &vinyl.Title, &vinyl.Releasedate, &vinyl.Price, &vinyl.Rating)
	if err != nil {
		return nil, err
	}
	return vinyl, nil
}

func updateVinylHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	vars := mux.Vars(r)
	idStr := vars["id"]

	vinylID, err := strconv.Atoi(idStr)

	var vinyl Vinyl
	err = json.NewDecoder(r.Body).Decode(&vinyl)

	UpdateVinyl(db, vinylID, vinyl.Title, vinyl.Artist, vinyl.Releasedate, vinyl.Price, vinyl.Rating)
	if err != nil {
		http.Error(w, "Vinyl not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "Vinyl updated successfully")
}

func UpdateVinyl(db *sql.DB, id int, title, artist string, releasedate, price int32, rating float32) error {
	query := "UPDATE vinyls SET title = ?, artist = ?, releasedate = ?, price = ?, rating = ? WHERE id = ?"
	_, err := db.Exec(query, title, artist, releasedate, price, rating, id)
	if err != nil {
		return err
	}
	return nil
}

func deleteVinylHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	vars := mux.Vars(r)
	idStr := vars["id"]

	vinylID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid 'id' parameter", http.StatusBadRequest)
		return
	}

	user := DeleteVinyl(db, vinylID)
	if err != nil {
		http.Error(w, "Vinyl not found", http.StatusNotFound)
		return
	}

	fmt.Fprintln(w, "Vinyl deleted successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func DeleteVinyl(db *sql.DB, id int) error {
	query := "DELETE FROM vinyls WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func handleJSONRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid JSON Format", http.StatusBadRequest)
		return
	}
	if requestBody.Message == "" {
		errorMessage := Response{
			Status:  "400",
			Message: "Invalid JSON message",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorMessage)
		return
	}

	fmt.Println("Recieved message: ", requestBody.Message)

	response := Response{
		Status:  "success",
		Message: "data successfully received  ",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

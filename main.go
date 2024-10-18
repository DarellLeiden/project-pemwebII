package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Struct untuk User
type User struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Email string `json:"email"`
}

// Fungsi untuk koneksi ke database
func connectDB() *sql.DB {
	connStr := "George:123456@tcp(localhost:3306)/project_pemwebii"
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	return db
}

// Fungsi untuk melakukan Basic Authentication
func basicAuth(next http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mendapatkan nilai dari header `Authorization`
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Memeriksa format header `Authorization`
		splitAuth := strings.SplitN(authHeader, " ", 2)
		if len(splitAuth) != 2 || splitAuth[0] != "Basic" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		// Mendekode nilai Base64 dari header `Authorization`
		decoded, err := base64.StdEncoding.DecodeString(splitAuth[1])
		if err != nil {
			http.Error(w, "Invalid Authorization credentials", http.StatusUnauthorized)
			return
		}

		// Memisahkan username dan password dari hasil decode
		pair := strings.SplitN(string(decoded), ":", 2)
		if len(pair) != 2 || pair[0] != username || pair[1] != password {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		// Jika authentication berhasil, lanjutkan ke handler berikutnya
		next.ServeHTTP(w, r)
	}
}

// Handler untuk endpoint root "/"
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Home Page! Please use /users endpoint for CRUD operations.")
}

// Handler untuk GET data dari tabel users
func getData(w http.ResponseWriter, r *http.Request) {
	db := connectDB()
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, email FROM users")
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Nama, &user.Email); err != nil {
			http.Error(w, "Error scanning data", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	json.NewEncoder(w).Encode(users)
}

// Handler untuk POST data ke tabel users
func createUser(w http.ResponseWriter, r *http.Request) {
	db := connectDB()
	defer db.Close()

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO users (nama, email) VALUES (?, ?)", user.Nama, user.Email)
	if err != nil {
		http.Error(w, "Error inserting data", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// Handler untuk PUT data (update) di tabel users
func updateUser(w http.ResponseWriter, r *http.Request) {
	db := connectDB()
	defer db.Close()

	// Ambil parameter ID dari URL
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	// Decode data JSON yang diterima dari request body
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	// Gunakan ID dari parameter URL untuk update
	_, err := db.Exec("UPDATE users SET nama=?, email=? WHERE id=?", user.Nama, user.Email, id)
	if err != nil {
		http.Error(w, "Error updating data", http.StatusInternalServerError)
		return
	}

	// Set ID yang di-update sesuai dengan parameter URL
	user.ID, _ = strconv.Atoi(id) // Konversi ID dari string ke int
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// Handler untuk DELETE data dari tabel users
func deleteUser(w http.ResponseWriter, r *http.Request) {
	db := connectDB()
	defer db.Close()

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		http.Error(w, "Error deleting data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User with ID %s has been deleted", id)
}

func main() {
	// Definisikan username dan password untuk Basic Authentication
	username := "admin"
	password := "password"

	// Daftarkan handler dengan Basic Authentication
	http.HandleFunc("/", basicAuth(homePage, username, password))               // Menangani endpoint "/" untuk menampilkan pesan di halaman utama
	http.HandleFunc("/users", basicAuth(getData, username, password))           // Menangani endpoint "/users" untuk GET data
	http.HandleFunc("/users/create", basicAuth(createUser, username, password)) // Menangani endpoint "/users/create" untuk POST data
	http.HandleFunc("/users/update", basicAuth(updateUser, username, password)) // Menangani endpoint "/users/update" untuk PUT data
	http.HandleFunc("/users/delete", basicAuth(deleteUser, username, password)) // Menangani endpoint "/users/delete" untuk DELETE data

	fmt.Println("Server is running on http://localhost:8080/")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}

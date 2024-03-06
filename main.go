package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	FirstName   string    `json:"firstname"`
	MiddleName  string    `json:"middlename,omitempty"`
	LastName    string    `json:"lastname"`
	Email       string    `json:"email"`
	Address     string    `json:"address,omitempty"`
	City        string    `json:"city,omitempty"`
	State       string    `json:"state,omitempty"`
	PostalCode  string    `json:"postal_code,omitempty"`
	Country     string    `json:"country,omitempty"`
	DateOfBirth time.Time `json:"date_of_birth,omitempty"`
	PhoneNumber string    `json:"phone_number,omitempty"`
	NIN         string    `json:"nin"`
	BVN         string    `json:"bvn"`
	AccountType string    `json:"account_type"`
}

type Account struct {
	AccountNumber   string  `json:"account_number"`
	UserID          int     `json:"user_id"`
	Tier            int     `json:"tier"`
	Type            string  `json:"type"`
	Currency        string  `json:"currency"`
	CanCredit       bool    `json:"can_credit"`
	CanDebit        bool    `json:"can_debit"`
	Balance         float64 `json:"balance"`
	Status          bool    `json:"status"`
	AdditionalField []byte  `json:"additional_field"`
}

var db *sql.DB

func init() {
	// Open database connection
	var err error
	db, err = sql.Open("mysql", "root:@tcp(localhost:3306)/cba")
	if err != nil {
		log.Fatal(err)
	}
	// Check database connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func createAccount(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user data
	if err := validateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Start database transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			return
		}
	}()

	// Create the user
	userID, err := CreateUser(tx, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accountNumber := generateAccountNumber()

	// Create the account data
	account := Account{
		AccountNumber:   accountNumber,
		UserID:          userID,
		Tier:            1,
		Type:            user.AccountType,
		Currency:        "NAIRA",
		CanCredit:       true,
		CanDebit:        true,
		Balance:         10000.00,
		Status:          true,
		AdditionalField: []byte("{}"),
	}

	// Create the account
	err = CreateAccount(tx, account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go SendAccountCreatedEmail(user, account)

	// Respond with success message and data
	response := map[string]interface{}{
		"status":  true,
		"message": "Account created successfully",
		"user":    user,
		"account": account,
	}
	jsonResponse(w, http.StatusOK, response)
}

func validateUser(user User) error {
	// Implement user data validation logic here
	// Return an error if validation fails, otherwise return nil
	return nil
}

func generateAccountNumber() string {
	// Implement account number generation logic here
	// Return a randomly generated account number
	return "1234567890" // Example account number
}

func CreateUser(tx *sql.Tx, user User) (int, error) {
	// Implement user creation logic here
	// Execute the SQL insert statement to create a new user
	// Assuming there is a table named "users" with appropriate columns
	// Assuming the database package is used to interact with the database
	result, err := tx.Exec("INSERT INTO users (firstname, middlename, lastname, email, address, city, state, postal_code, country, date_of_birth, phone_number, nin, bvn, account_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		user.FirstName, user.MiddleName, user.LastName, user.Email, user.Address, user.City, user.State, user.PostalCode, user.Country, user.DateOfBirth, user.PhoneNumber, user.NIN, user.BVN, user.AccountType)
	if err != nil {
		return 0, err
	}
	// Get the ID of the newly inserted user
	userID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(userID), nil
}

func CreateAccount(tx *sql.Tx, account Account) error {
	// Implement account creation logic here
	// Execute the SQL insert statement to create a new account
	// Assuming there is a table named "accounts" with appropriate columns
	// Assuming the database package is used to interact with the database
	_, err := tx.Exec("INSERT INTO accounts (account_number, user_id, tier, type, currency, can_credit, can_debit, balance, status, additional_field) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		account.AccountNumber, account.UserID, account.Tier, account.Type, account.Currency, account.CanCredit, account.CanDebit, account.Balance, account.Status, account.AdditionalField)
	return err
}

func SendAccountCreatedEmail(user User, account Account) {
	// Implement email sending logic here
	// This function can be run asynchronously using goroutines
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/createAccount", createAccount)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

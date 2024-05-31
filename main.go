package main

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

func main() {
	port := getPort()

	// Open database connection
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/bank_v1")
	if err != nil {
		return
	}
	defer db.Close() // Close the database connection when main() exits

	// Ping the database to check if the connection is successful
	if err := db.Ping(); err != nil {
		return
	}

	router := setupRouter(db)
	router.Run(":" + port)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	return port
}

func setupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("paginas/*.tmpl.html")
	router.Static("/static", "static")

	// Pass the database connection to each request's context using middleware
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Define routes
	router.GET("/", renderTemplate("login.tmpl.html")) // Login page
	router.POST("/login", handleLogin)
	router.GET("/transaction", renderTemplate("transaction.tmpl.html")) // Render transaction page
	router.POST("/create-transaction", handleTransfer)

	return router
}

func renderTemplate(templateName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, templateName, nil)
	}
}

func handleLogin(c *gin.Context) {
	// Get email and password from the request body
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Get the database connection from the context
	db := c.MustGet("db").(*sql.DB)

	// Query the database to check if the user exists and the password matches
	var userName string
	err := db.QueryRow("SELECT userName FROM userdetails WHERE userEmail = ? AND userPassword = ?", email, password).Scan(&userName)
	if err != nil {
		// Authentication failed
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var userID int
	query := db.QueryRow("SELECT userId FROM userdetails WHERE userName = ? AND userEmail = ? AND userPassword = ?", userName, email, password).Scan(&userID)
	if query != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	var (
		userEmail            string
		creditCardNumber     string
		currentBalance       float64
		creditCardExpiration string
		creditCardCVV        string
	)
	userData := db.QueryRow("SELECT userdetails.userEmail, useraccount.userAccountCurrentMoney, creditcards.cardNumber, creditcards.expirationDate, creditcards.cvv FROM userdetails JOIN useraccount ON userdetails.userId = useraccount.userId JOIN creditcards ON userdetails.userId = creditcards.userId WHERE userdetails.userId = ?", userID).Scan(&userEmail, &currentBalance, &creditCardNumber, &creditCardExpiration, &creditCardCVV)
	if userData != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Fetch the transaction history
	rows, err := db.Query("SELECT transactions.recipient_id, transactions.amount, transactions.description, transactions.transaction_date FROM transactions WHERE transactions.sender_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var (
			recipientId     int
			amount          float64
			description     string
			transactionDate string
		)
		err := rows.Scan(&recipientId, &amount, &description, &transactionDate)
		if err != nil {
			continue
		}

		var recipientName string
		err = db.QueryRow("SELECT userName FROM userdetails WHERE userId = ?", recipientId).Scan(&recipientName)
		if err != nil {
			continue
		}

		transaction := map[string]interface{}{
			"recipientName":   recipientName,
			"amount":          amount,
			"description":     description,
			"transactionDate": transactionDate,
		}
		transactions = append(transactions, transaction)
	}

	c.HTML(http.StatusOK, "homepage.tmpl.html", gin.H{
		"username":             userName,
		"useremail":            userEmail,
		"currentbalance":       currentBalance,
		"creditcardnumber":     creditCardNumber,
		"creditcardexpiration": creditCardExpiration,
		"creditcardcvv":        creditCardCVV,
		"transactions":         transactions,
	})
}

func handleTransfer(c *gin.Context) {
	senderID, err := strconv.Atoi(c.PostForm("senderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sender ID"})
		return
	}
	recipientID, err := strconv.Atoi(c.PostForm("recipientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
		return
	}
	amount, err := strconv.ParseFloat(c.PostForm("amount"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}
	description := c.PostForm("description")
	db := c.MustGet("db").(*sql.DB)

	err = transferMoney(senderID, recipientID, amount, description, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Transação efetuada com sucesso!"})
}

func transferMoney(senderID, recipientID int, amount float64, description string, db *sql.DB) error {
	// Register the transaction (the triggers will handle the balance updates)
	_, err := db.Exec("INSERT INTO transactions (sender_id, recipient_id, amount, description) VALUES (?, ?, ?, ?)",
		senderID, recipientID, amount, description)
	return err
}

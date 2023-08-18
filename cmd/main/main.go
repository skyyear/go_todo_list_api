package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Create a global variable for database connection
var db *sql.DB

type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Complete    bool      `json:"complete"`
	LastUpdated time.Time `json:"last_updated"`
}

// Retrieves all Todos
func getTodos(c *gin.Context) {

	var todoList []Todo

	sqlStatement := `SELECT * FROM todo`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Complete, &todo.LastUpdated)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		todoList = append(todoList, todo)
	}

	// Return Todos in Array
	c.JSON(http.StatusOK, todoList)
}

// Create a new todo
func createTodo(c *gin.Context) {

	sqlStatement := `
    INSERT INTO todo (title, complete)
    VALUES ($1, $2)
    RETURNING id, last_updated;`

	var newTodo Todo
	if err := c.ShouldBindJSON(&newTodo); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := db.QueryRow(sqlStatement, newTodo.Title, newTodo.Complete).Scan(&newTodo.ID, &newTodo.LastUpdated)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new todo successfully")
	c.JSON(http.StatusCreated, newTodo)
}

// Retreieve todo by id
func getTodoByID(c *gin.Context) {
	id := c.Param("id")

	sqlStatement := `
	SELECT * FROM todo 
	WHERE id = $1;`

	var todo Todo
	err := db.QueryRow(sqlStatement, id).Scan(&todo.ID, &todo.Title, &todo.Complete, &todo.LastUpdated)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Found match for todo with given id")

	// Returns todo based on id
	c.JSON(http.StatusOK, todo)
}

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

// Initialize database connection in main function
func main() {

	host := goDotEnvVariable("HOST")
	port := 2022
	user := goDotEnvVariable("DB_USER")
	password := goDotEnvVariable("DB_PASSWORD")
	dbname := goDotEnvVariable("DB_NAME")

	// create connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open database connection
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	// Register routes for CRUD operations
	router.GET("/todos", getTodos)        // Get all todos
	router.POST("/todos", createTodo)     // Create a new todo
	router.GET("/todos/:id", getTodoByID) // Get a single todo by ID

	router.Run("localhost:8080")
}

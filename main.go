package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		RETURNING id, last_updated;
	`

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
		WHERE id = $1;
	`

	var todo Todo
	err := db.QueryRow(sqlStatement, id).Scan(&todo.ID, &todo.Title, &todo.Complete, &todo.LastUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Todo with id %s not found", id)})
			return
		}
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Found match for todo with given id")

	// Returns todo based on id
	c.JSON(http.StatusOK, todo)
}

// Update existing Todo by ID
func updateTodoByID(c *gin.Context) {
	id := c.Param("id")

	var updatedTodo Todo
	if err := c.ShouldBindJSON(&updatedTodo); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sqlStatement := `
        UPDATE todo
        SET title = $1, complete = $2
        WHERE id = $3`

	res, err := db.Exec(sqlStatement, updatedTodo.Title, updatedTodo.Complete, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rows == 0 {
		// Send a not found response with an error message saying the given id doesn't exist
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("todo with id %s does not exist", id)})
	} else {
		// Send a JSON message response with the number of affected rows
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d row(s) updated", rows)})
	}
}

// Delete todo by ID
func deleteTodoByID(c *gin.Context) {
	id := c.Param("id")

	sqlStatement := `
        DELETE FROM todo
        WHERE id = $1`

	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if rows == 0 {
		// Send a not found response with an error message
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("todo with id %s does not exist", id)})
	} else {
		// Send a success response with a message
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d row(s) deleted", rows)})
	}
}

// Delete multiple todos by ids
func deleteTodosByIDs(c *gin.Context) {
	var ids []int
	err := c.BindJSON(&ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert the list of ids to a comma-separated string
	idStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")

	sqlStatement := `
		DELETE FROM todo 
		WHERE id IN (%s) 
		RETURNING id;
	`

	sql := fmt.Sprintf(sqlStatement, idStr)

	rows, err := db.Query(sql)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var deleted []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		deleted = append(deleted, id)
	}

	rowsAffected := len(deleted)

	// Find the difference between the original and deleted ids
	var notFound []int
	for _, id := range ids {
		if !contains(deleted, id) {
			notFound = append(notFound, id)
		}
	}

	// Return a success message with the number of rows deleted and which ids were not found
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d todos deleted", rowsAffected), "ids_not_found": notFound})
}

// Helper function to check if a slice contains an element
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Initialize database connection in main function
func main() {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// create connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	fmt.Println(psqlInfo)

	// Open database connection
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to open connection to Postgres")
		log.Fatal(err)
	}

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed test connection to Postgres")
		log.Fatal(err)
	}

	router := gin.Default()

	// Register routes for CRUD operations
	router.GET("/todos", getTodos)              // Get all todos
	router.POST("/todos", createTodo)           // Create a new todo
	router.GET("/todos/:id", getTodoByID)       // Get a single todo by ID
	router.PUT("/todos/:id", updateTodoByID)    // Update a todo by ID
	router.DELETE("/todos/:id", deleteTodoByID) // Delete a todo by ID
	router.DELETE("/todos", deleteTodosByIDs)   // Delete multiple todos by ids

	router.Run("0.0.0.0:8080")
}

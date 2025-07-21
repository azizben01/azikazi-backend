package functions

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/azikazi/azikazi/database"
	"github.com/gin-gonic/gin"
)



type Task struct {
	Task_id        int    `json:"task_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Category       string `json:"category"`
	Location       string `json:"location"`
	TimePreference string `json:"time_preference"`
	Price          int    `json:"price"`
	Status         string `json:"status"`
	PostedBy       int    `json:"posted_by"`
	AssignedTo     *int   `json:"assigned_to"`
	ExpiresAt      string `json:"expires_at"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	PostedByName   string `json:"posted_by_name"` // ✅ new
}

func PostTask(c *gin.Context) {
	var task Task

	// Parse request
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	userID := c.GetInt("user_id") // from auth middleware
	fmt.Println("Decoded userID:", userID)

	task.PostedBy = userID
task.Status = "open"

_, err := database.DB.Exec(`INSERT INTO task
(title, description, category, location, time_preference, price, status, posted_by) 
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
task.Title, task.Description, task.Category, task.Location,
task.TimePreference, task.Price, task.Status, task.PostedBy)


	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to post task"})
		fmt.Println("failed to post the task:", err)
		return
	}

	c.JSON(201, gin.H{"message": "Task posted successfully"})
}

func GetAllTasks(c *gin.Context) {
	rows, err := database.DB.Query(`
  SELECT 
    t.task_id, t.title, t.description, t.category, t.location,
    t.time_preference, t.price, t.status, u.name
  FROM task t
  JOIN users u ON t.posted_by = u.user_id
  WHERE t.status = 'open'
  ORDER BY t.created_at DESC
`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.Task_id, &t.Title, &t.Description, &t.Category, &t.Location, &t.TimePreference, &t.Price, &t.Status, &t.PostedByName,)
		if err != nil {
			fmt.Println("Scan error:", err)
		  }
		tasks = append(tasks, t)
	}

	c.JSON(200, tasks)
}



////// GetTaskByID returns a single task by its ID
func GetTaskByID(c *gin.Context) {
	// Get ID from the URL
	idParam := c.Param("id")
	taskID, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task Task
	// Query with JOIN to also get the name of the poster
	err = database.DB.QueryRow(`
		SELECT t.task_id, t.title, t.description, t.category, t.location, 
		       t.time_preference, t.price, t.status, u.name
		FROM task t
		JOIN users u ON t.posted_by = u.user_id
		WHERE t.task_id = $1
	`, taskID).Scan(
		&task.Task_id,
		&task.Title,
		&task.Description,
		&task.Category,
		&task.Location,
		&task.TimePreference,
		&task.Price,
		&task.Status,
		&task.PostedByName, // make sure Task struct has this field!
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

///// Update task by its ID
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var t Task

	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	_, err := database.DB.Exec(`
		UPDATE task
		SET title=$1, description=$2, category=$3, location=$4, time_preference=$5, price=$6, updated_at=NOW()
		WHERE task_id = $7
	`, t.Title, t.Description, t.Category, t.Location, t.TimePreference, t.Price, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(200, gin.H{"message": "Task updated"})
}



///// tasks posted by the logged in user
func GetMyTasks(c *gin.Context) {
	userID := c.GetInt("user_id")
	rows, err := database.DB.Query(`
	  SELECT task_id, title, location, status, description, category, time_preference, price
	  FROM task
	  WHERE posted_by = $1
	  ORDER BY created_at DESC
	`, userID)
	if err != nil {
	  c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
	  return
	}
  
	var tasks []Task
	for rows.Next() {
	  var t Task
	  rows.Scan(&t.Task_id, &t.Title, &t.Location, &t.Status, &t.Description, &t.Category, &t.TimePreference, &t.Price)
	  tasks = append(tasks, t)
	}
  
	c.JSON(200, tasks)
  }
  

  ///// The function to delete a task 

func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetInt("user_id")
  
	result, err := database.DB.Exec(
	  "DELETE FROM task WHERE task_id = $1 AND posted_by = $2",
	  taskID, userID,
	)
  
	if err != nil {
	  c.JSON(500, gin.H{"error": "Failed to delete task"})
	  return
	}
  
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
	  c.JSON(404, gin.H{"error": "Task not found or you are not the owner"})
	  return
	}
  
	c.JSON(200, gin.H{"message": "Task deleted successfully"})
  }
  
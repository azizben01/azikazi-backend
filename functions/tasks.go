package functions

import (
	"fmt"

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

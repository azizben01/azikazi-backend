package functions

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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
	IsDeleted      bool   `json:"is_deleted"`
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
  WHERE t.status = 'open' AND t.isDeleted = FALSE AND u.isDeleted = FALSE
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
		WHERE t.task_id = $1 AND t.isDeleted = FALSE AND u.isDeleted = FALSE
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
		WHERE task_id = $7 AND isDeleted = FALSE
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
	  WHERE posted_by = $1 AND isDeleted = FALSE
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
	  "UPDATE task SET isDeleted = TRUE WHERE task_id = $1 AND posted_by = $2",
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
  
  /// The function to apply to a task
  
  func ApplyToTask(c *gin.Context) {
    taskIDParam := c.Param("id")

    // Parse task ID
    taskID, err := strconv.Atoi(taskIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
        return
    }

    // Get user ID from context
    userIDInterface, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }
    userID := userIDInterface.(int)

    // Parse optional message
    var body struct {
        Message string `json:"message"`
    }
    if err := c.BindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    // Prevent duplicate application
    var count int
    err = database.DB.QueryRow(`
        SELECT COUNT(*) FROM task_applications WHERE task_id = $1 AND applicant_id = $2 AND isDeleted = FALSE
    `, taskID, userID).Scan(&count)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing application"})
        return
    }
    if count > 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "You have already applied to this task"})
        return
    }

    // Insert into task_applications
    _, err = database.DB.Exec(`
        INSERT INTO task_applications (task_id, applicant_id, message)
        VALUES ($1, $2, $3)
    `, taskID, userID, body.Message)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply for task"})
        return 
    }

    c.JSON(http.StatusOK, gin.H{"message": "Application submitted successfully"})
}


//////// Function that fetches the application received for a specifique task

func GetTaskApplications(c *gin.Context) {
    taskIDStr := c.Param("id")
    taskID, err := strconv.Atoi(taskIDStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid task ID"})
		log.Println("Error fetching applications 1:", err)
        return
    }

    // Get the user ID from the token (middleware)
    userID := c.GetInt("user_id")

    // Check that the user is the owner of the task
    var postedBy int
    err = database.DB.QueryRow("SELECT posted_by FROM task WHERE task_id=$1 AND isDeleted = FALSE", taskID).Scan(&postedBy)
    if err != nil {
        c.JSON(404, gin.H{"error": "Task not found"})
		log.Println("Error fetching applications 2:", err)
        return
    }

    if postedBy != userID {
        c.JSON(403, gin.H{"error": "Unauthorized: You do not own this task"})
		log.Println("Error fetching applications 3:", err)
        return
    }

    // Fetch applications with user info
    rows, err := database.DB.Query(`
        SELECT a.application_id, a.message, a.created_at,
               u.user_id, u.name, u.email, a.status
        FROM task_applications a
        JOIN users u ON a.applicant_id = u.user_id
        WHERE a.task_id = $1 AND a.isDeleted = FALSE AND u.isDeleted = FALSE
        ORDER BY a.created_at DESC
    `, taskID)

    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch applications"})
		log.Println("Error fetching applications 4:", err)

        return
    }

    defer rows.Close()

    var applications []gin.H
    for rows.Next() {
        var appID, userID int
        var message, name, email, status string
        var appliedAt time.Time

        if err := rows.Scan(&appID, &message, &appliedAt, &userID, &name, &email, &status); err != nil {
            continue
        }

        applications = append(applications, gin.H{
            "application_id": appID,
            "user_id":        userID,
            "user_name":      name,
            "user_email":     email,
            "message":        message,
            "applied_at":     appliedAt,
			"status":         status,
        })
    }

    c.JSON(200, applications)
}

// Function with the logic to accept a task and now hidding from others
func AcceptApplication(c *gin.Context) {
    appIDStr := c.Param("application_id")
    appID, err := strconv.Atoi(appIDStr)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid application ID"})
        return
    }

    userID := c.GetInt("user_id")

    // Fetch task_id and applicant_id of this application
    var taskID, applicantID int
    err = database.DB.QueryRow(`SELECT task_id, applicant_id FROM task_applications WHERE application_id=$1 AND isDeleted = FALSE`, appID).Scan(&taskID, &applicantID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Application not found"})
        return
    }

    // Verify user owns the task
    var postedBy int
    err = database.DB.QueryRow(`SELECT posted_by FROM task WHERE task_id=$1 AND isDeleted = FALSE`, taskID).Scan(&postedBy)
	if err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		log.Println("Error accepting application - task fetch:", err)
		return
	}
	
	if postedBy != userID {
		c.JSON(403, gin.H{"error": "Unauthorized: You do not own this task"})
		log.Printf("Authorization error: task postedBy=%d, your userID=%d\n", postedBy, userID)
		return
	}
	
    tx, err := database.DB.Begin()
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to start transaction"})
        return
    }

    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

    // 1. Accept the selected application
    _, err = tx.Exec(`UPDATE task_applications SET status='accepted' WHERE application_id=$1`, appID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to update application"})
		log.Println("Error accepting application 1:", err)
        return
    }

    // 2. Reject other applications for same task
    _, err = tx.Exec(`UPDATE task_applications SET status='rejected' WHERE task_id=$1 AND application_id != $2`, taskID, appID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to reject other applications"})
		log.Println("Error accepting application 2:", err)
        return
    }

    // 3. Update the task with assigned_to and status
    _, err = tx.Exec(`UPDATE task SET assigned_to=$1, status='accepted' WHERE task_id=$2`, applicantID, taskID)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to update task"})
		log.Println("Error accepting application 3:", err)
        return
    }

    c.JSON(200, gin.H{"message": "Application accepted and task updated"})
}

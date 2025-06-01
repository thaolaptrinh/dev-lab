package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Age   int    `json:"age,omitempty"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var users = []User{
	{
		ID:        1,
		Name:      "Nguyen Van Thao",
		Email:     "thao@example.com",
		Age:       25,
		CreatedAt: time.Now(),
	},
	{
		ID:        2,
		Name:      "Tran Thi Mai",
		Email:     "mai@example.com",
		Age:       30,
		CreatedAt: time.Now(),
	},
}

var nextID = 3

// Middleware để log requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s - Started", r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		log.Printf("%s %s - Completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Helper function để gửi JSON response
func sendJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Helper function để tìm user theo ID
func findUserByID(id int) (*User, int) {
	for i, user := range users {
		if user.ID == id {
			return &user, i
		}
	}
	return nil, -1
}

func main() {
	mux := http.NewServeMux()

	// Home endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			sendJSONResponse(w, http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Endpoint not found",
			})
			return
		}

		sendJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Welcome to Go REST API",
			Data:    "Server is running on port 8080",
		})
	})

	// GET /users - Lấy danh sách tất cả users
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sendJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "Users retrieved successfully",
				Data:    users,
			})

		case http.MethodPost:
			// POST /users - Tạo user mới
			var req CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				sendJSONResponse(w, http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "Invalid JSON format",
				})
				return
			}

			// Validation
			if req.Name == "" || req.Email == "" {
				sendJSONResponse(w, http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "Name and email are required",
				})
				return
			}

			// Kiểm tra email đã tồn tại chưa
			for _, user := range users {
				if user.Email == req.Email {
					sendJSONResponse(w, http.StatusConflict, APIResponse{
						Success: false,
						Message: "Email already exists",
					})
					return
				}
			}

			// Tạo user mới
			newUser := User{
				ID:        nextID,
				Name:      req.Name,
				Email:     req.Email,
				Age:       req.Age,
				CreatedAt: time.Now(),
			}
			users = append(users, newUser)
			nextID++

			sendJSONResponse(w, http.StatusCreated, APIResponse{
				Success: true,
				Message: "User created successfully",
				Data:    newUser,
			})

		default:
			sendJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
		}
	})

	// GET/PUT/DELETE /users/{id} - Thao tác với một user cụ thể
	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/users/")
		id, err := strconv.Atoi(idStr)

		if err != nil {
			sendJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid user ID",
			})
			return
		}

		switch r.Method {
		case http.MethodGet:
			// GET /users/{id} - Lấy thông tin một user
			user, _ := findUserByID(id)
			if user == nil {
				sendJSONResponse(w, http.StatusNotFound, APIResponse{
					Success: false,
					Message: "User not found",
				})
				return
			}

			sendJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "User retrieved successfully",
				Data:    user,
			})

		case http.MethodPut:
			// PUT /users/{id} - Cập nhật thông tin user
			user, index := findUserByID(id)
			if user == nil {
				sendJSONResponse(w, http.StatusNotFound, APIResponse{
					Success: false,
					Message: "User not found",
				})
				return
			}

			var req UpdateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				sendJSONResponse(w, http.StatusBadRequest, APIResponse{
					Success: false,
					Message: "Invalid JSON format",
				})
				return
			}

			// Cập nhật thông tin (chỉ cập nhật các field không rỗng)
			if req.Name != "" {
				users[index].Name = req.Name
			}
			if req.Email != "" {
				// Kiểm tra email trùng với user khác
				for i, u := range users {
					if i != index && u.Email == req.Email {
						sendJSONResponse(w, http.StatusConflict, APIResponse{
							Success: false,
							Message: "Email already exists",
						})
						return
					}
				}
				users[index].Email = req.Email
			}
			if req.Age > 0 {
				users[index].Age = req.Age
			}

			sendJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "User updated successfully",
				Data:    users[index],
			})

		case http.MethodDelete:
			// DELETE /users/{id} - Xóa user
			_, index := findUserByID(id)
			if index == -1 {
				sendJSONResponse(w, http.StatusNotFound, APIResponse{
					Success: false,
					Message: "User not found",
				})
				return
			}

			// Xóa user khỏi slice
			users = append(users[:index], users[index+1:]...)

			sendJSONResponse(w, http.StatusOK, APIResponse{
				Success: true,
				Message: "User deleted successfully",
			})

		default:
			sendJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
		}
	})

	// GET /users/search?name=xxx - Search users by name
	mux.HandleFunc("/users/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			sendJSONResponse(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Name parameter is required",
			})
			return
		}

		var foundUsers []User
		for _, user := range users {
			if strings.Contains(strings.ToLower(user.Name), strings.ToLower(name)) {
				foundUsers = append(foundUsers, user)
			}
		}

		sendJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: fmt.Sprintf("Found %d users", len(foundUsers)),
			Data:    foundUsers,
		})
	})

	// GET /stats - Thống kê
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendJSONResponse(w, http.StatusMethodNotAllowed, APIResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		totalUsers := len(users)
		var totalAge int
		for _, user := range users {
			totalAge += user.Age
		}

		var averageAge float64
		if totalUsers > 0 {
			averageAge = float64(totalAge) / float64(totalUsers)
		}

		stats := map[string]interface{}{
			"total_users":  totalUsers,
			"average_age":  averageAge,
			"server_time":  time.Now().Format("2006-01-02 15:04:05"),
		}

		sendJSONResponse(w, http.StatusOK, APIResponse{
			Success: true,
			Message: "Statistics retrieved successfully",
			Data:    stats,
		})
	})

	// Áp dụng middleware logging
	handler := loggingMiddleware(mux)

	fmt.Println("🚀 Server is starting on port 8080...")
	fmt.Println("📋 Available endpoints:")
	fmt.Println("  GET    /              - Welcome message")
	fmt.Println("  GET    /users         - Get all users")
	fmt.Println("  POST   /users         - Create new user")
	fmt.Println("  GET    /users/{id}    - Get user by ID")
	fmt.Println("  PUT    /users/{id}    - Update user by ID")
	fmt.Println("  DELETE /users/{id}    - Delete user by ID")
	fmt.Println("  GET    /users/search  - Search users by name")
	fmt.Println("  GET    /stats         - Get statistics")
	fmt.Println()

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
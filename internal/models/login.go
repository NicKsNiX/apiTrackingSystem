package models

// LoginRequest - struct to receive username and password from client
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse - struct to send response back to client
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    User   `json:"data"`
}

// User - struct to represent user data
type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// ExternalUser represents the structure returned by the external auth service
type ExternalUser struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Position       string `json:"position"`
	ExtensionName  string `json:"extensionName"`
	Department     string `json:"department"`
	DepartmentCode string `json:"departmentNumber"`
	Division       string `json:"division"`
	Name           string `json:"name"`
	Surname        string `json:"surname"`
	EmployeeID     string `json:"employeeID"`
}

// ExternalLoginResponse is the top-level response from the external auth service
type ExternalLoginResponse struct {
	Message string       `json:"message"`
	User    ExternalUser `json:"user"`
}

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
	"gorm.io/gorm"

	"github.com/MoSed3/otp-server/internal/config"
	"github.com/MoSed3/otp-server/internal/db"
	"github.com/MoSed3/otp-server/internal/models"
	"github.com/MoSed3/otp-server/internal/repository"
)

// Helper function to securely prompt for password
func promptForPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("error reading password: %v", err)
	}
	fmt.Println() // Newline after password input
	return string(bytePassword), nil
}

// Helper function to prompt for string input, allowing empty to skip
func promptForInput(prompt string, defaultValue string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (current: %s, leave blank to keep current): ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %v", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue, nil
	}
	return input, nil
}

func handleCreate(tx *gorm.DB, adminRepo repository.Admin, args []string) {
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createUsernameFlag := createCmd.String("username", "", "Username for the new admin")
	createCmd.StringVar(createUsernameFlag, "u", "", "Username for the new admin (shorthand)")
	createRoleFlag := createCmd.String("role", "", fmt.Sprintf("Role for the new admin (available roles: %s, %s, %s)",
		models.RoleSuperAdmin.String(), models.RoleSudoAdmin.String(), models.RoleVisitorAdmin.String()))
	createCmd.StringVar(createRoleFlag, "r", "", "Role for the new admin (shorthand)")

	createCmd.Parse(args)

	username := *createUsernameFlag
	if username == "" {
		var err error
		username, err = promptForInput("Enter Username", "")
		if err != nil {
			log.Fatalf("Error getting username: %v", err)
		}
		if username == "" {
			log.Fatal("Username cannot be empty")
		}
	}

	password, err := promptForPassword("Enter Password: ")
	if err != nil {
		log.Fatalf("Error getting password: %v", err)
	}
	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	roleStr := *createRoleFlag
	if roleStr == "" {
		var err error
		roleStr, err = promptForInput(fmt.Sprintf("Enter Role (available roles: %s, %s, %s)",
			models.RoleSuperAdmin.String(), models.RoleSudoAdmin.String(), models.RoleVisitorAdmin.String()), models.RoleVisitorAdmin.String())
		if err != nil {
			log.Fatalf("Error getting role: %v", err)
		}
	}

	role, err := models.ParseAdminRole(roleStr)
	if err != nil {
		log.Fatalf("Invalid role: %s", roleStr)
	}

	admin, err := adminRepo.Create(tx, username, password, role)
	if err != nil {
		log.Fatalf("Error creating admin: %v", err)
	}
	fmt.Printf("Admin created successfully: ID=%d, Username=%s, Role: %s\n", admin.ID, admin.Username, admin.Role)
}

func handleUpdate(tx *gorm.DB, adminRepo repository.Admin, args []string) {
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateID := updateCmd.Uint("id", 0, "ID of the admin to update")
	updateCmd.UintVar(updateID, "i", 0, "ID of the admin to update (shorthand)")
	updateUsernameFlag := updateCmd.String("username", "", "New username for the admin (optional)")
	updateCmd.StringVar(updateUsernameFlag, "u", "", "New username for the admin (shorthand, optional)")
	updateRoleFlag := updateCmd.String("role", "", fmt.Sprintf("New role for the admin (optional, available roles: %s, %s, %s)", models.RoleSuperAdmin.String(), models.RoleSudoAdmin.String(), models.RoleVisitorAdmin.String()))
	updateCmd.StringVar(updateRoleFlag, "r", "", "New role for the admin (shorthand, optional)")

	updateCmd.Parse(args)
	if *updateID == 0 {
		updateCmd.PrintDefaults()
		os.Exit(1)
	}

	admin, err := adminRepo.GetByID(tx, *updateID)
	if err != nil {
		log.Fatalf("Admin with ID %d not found: %v", *updateID, err)
	}

	updates := repository.AdminUpdate{}

	// Handle username update
	username := *updateUsernameFlag
	if username == "" {
		username, err = promptForInput("Enter New Username", admin.Username)
		if err != nil {
			log.Fatalf("Error getting new username: %v", err)
		}
	}
	if username != admin.Username {
		updates.Username = &username
	}

	// Handle password update
	newPassword, err := promptForPassword("Enter New Password (leave blank to keep current): ")
	if err != nil {
		log.Fatalf("Error getting new password: %v", err)
	}
	if newPassword != "" {
		updates.NewPassword = &newPassword
	}

	// Handle role update
	roleStr := *updateRoleFlag
	if roleStr == "" {
		roleStr, err = promptForInput(fmt.Sprintf("Enter New Role (available roles: %s, %s, %s)",
			models.RoleSuperAdmin.String(), models.RoleSudoAdmin.String(), models.RoleVisitorAdmin.String()), admin.Role.String())
		if err != nil {
			log.Fatalf("Error getting new role: %v", err)
		}
	}
	if roleStr != admin.Role.String() {
		role, err := models.ParseAdminRole(roleStr)
		if err != nil {
			log.Fatalf("Invalid role: %s", roleStr)
		}
		updates.Role = &role
	}

	err = adminRepo.Update(tx, *updateID, updates)
	if err != nil {
		log.Fatalf("Error updating admin: %v", err)
	}
	fmt.Printf("Admin ID %d updated successfully\n", *updateID)
}

func handleDelete(tx *gorm.DB, adminRepo repository.Admin, args []string) {
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteCmd.Uint("id", 0, "ID of the admin to delete")

	deleteCmd.Parse(args)
	if *deleteID == 0 {
		deleteCmd.PrintDefaults()
		os.Exit(1)
	}

	err := adminRepo.Delete(tx, *deleteID)
	if err != nil {
		log.Fatalf("Error deleting admin: %v", err)
	}
	fmt.Printf("Admin ID %d deleted successfully\n", *deleteID)
}

func handleList(tx *gorm.DB, adminRepo repository.Admin) {
	admins, err := adminRepo.ListAll(tx)
	if err != nil {
		log.Fatalf("Error listing admins: %v", err)
	}
	if len(admins) == 0 {
		fmt.Println("No admin users found.")
		return
	}
	fmt.Println("Admin Users:")

	// Calculate maximum widths for columns
	maxIDLen := len("ID")
	maxUsernameLen := len("Username")
	maxRoleLen := len("Role")

	for _, admin := range admins {
		if idLen := len(fmt.Sprintf("%d", admin.ID)); idLen > maxIDLen {
			maxIDLen = idLen
		}
		if usernameLen := len(admin.Username); usernameLen > maxUsernameLen {
			maxUsernameLen = usernameLen
		}
		if roleLen := len(admin.Role.String()); roleLen > maxRoleLen {
			maxRoleLen = roleLen
		}
	}

	// Print header
	fmt.Printf("%-*s  %-*s  %-*s\n", maxIDLen, "ID", maxUsernameLen, "Username", maxRoleLen, "Role")
	fmt.Printf("%s  %s  %s\n",
		generateDash(maxIDLen),
		generateDash(maxUsernameLen),
		generateDash(maxRoleLen))

	// Print admin data
	for _, admin := range admins {
		fmt.Printf("%-*d  %-*s  %-*s\n", maxIDLen, admin.ID, maxUsernameLen, admin.Username, maxRoleLen, admin.Role.String())
	}
}

// Helper function to generate a string of dashes for formatting
func generateDash(length int) string {
	s := ""
	for i := 0; i < length; i++ {
		s += "-"
	}
	return s
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg := config.LoadConfig()
	database, err := db.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx := database.GetTransaction(ctx)
	adminRepo := repository.NewAdmin()

	switch os.Args[1] {
	case "create":
		handleCreate(tx, adminRepo, os.Args[2:])
	case "update":
		handleUpdate(tx, adminRepo, os.Args[2:])
	case "delete":
		handleDelete(tx, adminRepo, os.Args[2:])
	case "list":
		handleList(tx, adminRepo)
	case "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err := tx.Commit().Error; err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage: admin <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  create    Create a new admin user. Use 'admin create -h' for more details.")
	fmt.Println("  update    Update an existing admin user. Use 'admin update -h' for more details.")
	fmt.Println("  delete    Delete an admin user. Use 'admin delete -h' for more details.")
	fmt.Println("  list      List all admin users. Use 'admin list -h' for more details.")
	fmt.Println("  help      Display this help message.")
	fmt.Println("\nTo get help for a specific command, use: admin <command> -h")
}

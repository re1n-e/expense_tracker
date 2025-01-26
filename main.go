package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const filename = "expense.json"

func about() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("add\t\t\tUsers can add an expense with a description and amount.")
	fmt.Println("update\t\t\tUsers can update an expense.")
	fmt.Println("delete\t\t\tUsers can delete an expense.")
	fmt.Println("summary\t\t\tUsers can view all expenses.")
	fmt.Println("month --m\t\tUsers can view a summary of expenses for a specific month (of current year).")
	fmt.Println("Flags\t\t\tUsers can use flags to expand the command")
	fmt.Println("Usage: summary\t\t--Flag")
	fmt.Println("summary --month\t\tUsers can view a summary of all expenses.")
}

type expense struct {
	ID          int     `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

func readFile() ([]expense, error) {
	data, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	fileContent, err := io.ReadAll(data)
	if err != nil {
		return nil, err
	}

	if len(fileContent) == 0 {
		return []expense{}, nil
	}

	var expenses []expense
	if err := json.Unmarshal(fileContent, &expenses); err != nil {
		return nil, err
	}

	return expenses, nil
}

func writeFile(expenses []expense) error {
	data, err := json.MarshalIndent(expenses, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func add(description string, amount float64) error {
	expenses, err := readFile()
	if err != nil {
		return err
	}

	newExpense := expense{
		ID:          getID(expenses),
		Date:        time.Now().Format("2006-01-02"),
		Description: description,
		Amount:      amount,
	}

	expenses = append(expenses, newExpense)

	if err := writeFile(expenses); err != nil {
		return err
	}

	fmt.Printf("Expense added successfully (ID: %d)\n", newExpense.ID)
	return nil
}

func update(id int, description string, amount float64) error {
	expenses, err := readFile()
	if err != nil {
		return err
	}

	for i, t := range expenses {
		if t.ID == id {
			expenses[i].Description = description
			expenses[i].Amount = amount

			if err := writeFile(expenses); err != nil {
				return err
			}

			fmt.Printf("Expense %d updated successfully\n", id)
			return nil
		}
	}

	return fmt.Errorf("expense with ID %d not found", id)
}

func delete(id int) error {
	expenses, err := readFile()
	if err != nil {
		return err
	}

	for i, t := range expenses {
		if t.ID == id {
			expenses = append(expenses[:i], expenses[i+1:]...)
			if err := writeFile(expenses); err != nil {
				return err
			}
			fmt.Printf("Expense %d deleted successfully\n", id)
			return nil
		}
	}
	return fmt.Errorf("expense with id %d doesn't exist", id)
}

func summary(flag string, month int) error {
	expenses, err := readFile()
	if err != nil {
		return err
	}

	fmt.Printf("%-5s %-10s %-30s %-20s\n", "ID", "Date", "Description", "Amount")
	fmt.Println(strings.Repeat("-", 85))

	var total float64
	for _, t := range expenses {
		expenseDate, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return err
		}

		switch flag {
		case "--month":
			if expenseDate.Month() == time.Month(month) && expenseDate.Year() == time.Now().Year() {
				fmt.Printf("%-5d %-10s %-30s %-20.2f\n", t.ID, t.Date, t.Description, t.Amount)
				total += t.Amount
			}
		default:
			fmt.Printf("%-5d %-10s %-30s %-20.2f\n", t.ID, t.Date, t.Description, t.Amount)
			total += t.Amount
		}
	}

	fmt.Println(strings.Repeat("-", 85))
	fmt.Printf("Total: %.2f\n", total)
	return nil
}

func getID(expenses []expense) int {
	if len(expenses) == 0 {
		return 1
	}
	return expenses[len(expenses)-1].ID + 1
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: main <command>\nUse: \"main about\" to know all commands")
	}

	switch os.Args[1] {
	case "about":
		about()
	case "add":
		if len(os.Args) != 6 {
			log.Fatalf("Usage: main add --description <description> --amount <amount>")
		}
		if strings.HasPrefix(os.Args[2], "--") && strings.HasPrefix(os.Args[4], "--") {
			amount, err := strconv.ParseFloat(os.Args[5], 64)
			if err != nil {
				log.Fatalf("Err: %s. The amount is not a valid value", os.Args[5])
			}
			if err := add(os.Args[3], amount); err != nil {
				log.Fatalf("Failed to add.\nError: %v", err)
			}
		} else {
			log.Fatalf("Usage: main add --description <description> --amount <amount>")
		}
	case "update":
		if len(os.Args) != 7 {
			log.Fatalf("Usage: main update <id> --description <description> --amount <amount>")
		}
		if strings.HasPrefix(os.Args[3], "--") && strings.HasPrefix(os.Args[5], "--") {
			id, err := strconv.ParseInt(os.Args[2], 10, 64)
			if err != nil {
				log.Fatalf("Err: %s. The ID is not a valid value", os.Args[2])
			}
			amount, err := strconv.ParseFloat(os.Args[6], 64)
			if err != nil {
				log.Fatalf("Err: %s. The amount is not a valid value", os.Args[5])
			}
			if err := update(int(id), os.Args[4], amount); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Usage: main update <id> --description <description> --amount <amount>")
		}
	case "delete":
		if len(os.Args) != 3 {
			log.Fatalf("Usage: main delete <id>")
		}
		id, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			log.Fatalf("Err: %s. The ID is not a valid value", os.Args[2])
		}
		if err := delete(int(id)); err != nil {
			log.Fatal(err)
		}
	case "summary":
		if len(os.Args) == 2 {
			if err := summary("", 0); err != nil {
				log.Fatal(err)
			}
		} else if len(os.Args) == 4 && os.Args[2] == "--month" {
			month, err := strconv.Atoi(os.Args[3])
			if err != nil || month < 1 || month > 12 {
				log.Fatalf("Invalid month. Please use a number between 1 and 12")
			}
			if err := summary("--month", month); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Usage: main summary or main summary --month <month number>")
		}
	default:
		log.Fatalf("Unknown command. Use \"main about\" to see available commands.")
	}
}

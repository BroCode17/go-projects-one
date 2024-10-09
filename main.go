package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Priority int

const FILENAME = "db.json"

type Task struct {
	Text string `json:"name"`
	ID   int    `json:"id"`
	Priority Priority `json:"priority"`
	DueDate time.Time `json:"due_date"`
	Category string `json:"category"`
	IsComplete bool `json:"is_complete"`
}

const (
	Low Priority = iota
	Medium
	High
)

type TodoList struct {
	Tasks []Task
}

func main() {
	var todos TodoList

	//load todos
	if err := loadTodo(&todos); err != nil {
		fmt.Println("Could not load todo from file")
	}
	//command line flags
	addFlag := flag.String("add", "", "Add task to list")
	deleteFlag := flag.Int("delete", 0, "Delete item from todo")
	listFlag := flag.Bool("list", false, "List all todos")
    completeFalg  := flag.Int("complete", 0, "Filter tasks by category")
	filterFlag := flag.String("filter", "", "Filter tasks by category")
	sortFlag := flag.String("sort","", "Sort tasks by 'priority', 'due' or 'category'")
	interactiveFlag := flag.Bool("interactive", false, "Enter interactive mode")

	flag.Parse()

	if *addFlag != "" {
		add(&todos, *addFlag)
	} else if *deleteFlag != 0 {
		remove(&todos, *deleteFlag)
	} else if *listFlag {
		list(todos, *filterFlag, *sortFlag)
	} else if *completeFalg != 0 {
		markTaskComplete(&todos, *completeFalg)
	}else if *interactiveFlag {
		interactiveMode(&todos)
	}else {
		fmt.Println("Please provide a valid command. Use -h for help.")
	}
}

// load
func loadTodo(todo *TodoList) error {
	//read todo from file
	data, err := ioutil.ReadFile(FILENAME)

	if err != nil {
		//check if file does not exist
		//os error
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Unmarsh json data to object
	// return nil or error
	return json.Unmarshal(data, todo)
}

// save
func save(todo TodoList) error {
	//marsh todo
	//Convert todo to json object
	data, err := json.Marshal(todo)
	if err != nil {
		return err
	}
	//return err or nil
	return ioutil.WriteFile(FILENAME, data, 0644)
}

// add
func add(todos *TodoList, text string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter priority (0: Low, 1: Medium, 2: High): ")
	priorityString,_ := reader.ReadString('\n')
	priority, _ := strconv.Atoi(strings.TrimSpace(priorityString))

	fmt.Print("Enter due date (YYYY-MM-DD): ")
	dueDateStr, _ := reader.ReadString('\n')
	dueDate, _ := time.Parse("2006-01-02", strings.TrimSpace(dueDateStr))

	fmt.Print("Enter category: ")
	category, _ := reader.ReadString('\n')
	category = strings.TrimSpace(category)

	// Generating new id
	newID := 1
	if len(todos.Tasks) > 0 {
		//Get the last task id
		newID = todos.Tasks[len(todos.Tasks)-1].ID + 1
	}

	//append new task to slice
	todos.Tasks = append(todos.Tasks, Task{
		ID: newID, 
		Text: text,
		Priority: Priority(priority),
		DueDate: dueDate,
		Category: category,
		IsComplete: false,
	})

	//save
	if err := save(*todos); err != nil {
		fmt.Println("Error saving tasks: ", err)
		return
	}

	fmt.Printf("Task added: %s\n", text)
}

// remove
func remove(todos *TodoList, id int) {
	for i, task := range todos.Tasks {
		if task.ID == id {
			//create new slice with the deleted task
			todos.Tasks = append(todos.Tasks[:i], todos.Tasks[i+1:]...)
			//save
			if err := save(*todos); err != nil {
				fmt.Println("Error occured while saving: ", err)
				return
			}
			fmt.Printf("Task removed: %d\n", id)
			return
		}

	}
	fmt.Printf("Task not found: %d\n", id)
}

// list
func list(todos TodoList, filterCategory, sortBy string) {
	if len(todos.Tasks) == 0 {
		fmt.Println("No task")
		return
	}

	filteredTasks := todos.Tasks
	
	if filterCategory != "" {
		var tasks []Task
		for _, task := range todos.Tasks {
			if task.Category == filterCategory{
				tasks = append(tasks, task)
			}
		}
		filteredTasks = tasks
	}

	switch sortBy {
	case "priority":
		sort.Slice(filteredTasks, func(i, j int) bool {
			return filteredTasks[i].Priority > filteredTasks[j].Priority
		})
	case "due":
		sort.Slice(filteredTasks, func(i, j int) bool {
			return filteredTasks[i].DueDate.Before(filteredTasks[j].DueDate)
		})
	case "category":
		sort.Slice(filteredTasks, func(i, j int) bool {
			return filteredTasks[i].Category < filteredTasks[j].Category
		})
	}

	for _, task := range filteredTasks {
		status := " "
		if task.IsComplete {
			status = "✓"
		}
		fmt.Printf("[%s] %d: %s (Priority: %d, Due: %s, Category: %s)\n",
		 status, task.ID, task.Text, task.Priority, task.DueDate.Format("2006-01-02"), task.Category)
	}
 
}

//mark as complete
func markTaskComplete(todos *TodoList, id int)  {
	for i, task := range todos.Tasks {
		if task.ID == id {
			todos.Tasks[i].IsComplete = true
			if err := save(*todos); err != nil {
				fmt.Println("Error saving tasks: ",err)
				return
			}
			fmt.Println("Task marked as complete: ", id)
			return
		}
	}
}

// interactive mode
func interactiveMode(todos *TodoList){
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nInteractive Mode: ")
		fmt.Println("1. Add task")
		fmt.Println("2. Remove task")
		fmt.Println("3. List tasks")
		fmt.Println("4. Mark task as complete")
		fmt.Println("5: Exit")
		fmt.Print("Enter your choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter task description: ")
			text, _ := reader.ReadString('\n')
			add(todos, strings.TrimSpace(text))
		case "2":
			fmt.Print("Enter task ID to remove: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			remove(todos, id)
		case "3":
			fmt.Print("Enter filter category (or leave empty): ")
			filter, _ := reader.ReadString('\n')
			filter = strings.TrimSpace(filter)
			fmt.Print("Enter sort method (priority/due/category, or leave emtpy): ")
			sort, _ := reader.ReadString('\n')
			sort = strings.TrimSpace(sort)
			list(*todos, filter, sort)
		case "4":
			fmt.Print("Enter task ID to mark as complete: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			markTaskComplete(todos, id)
		case "5":
			fmt.Println("Exiting interactive mode.")
			return
		default:
			fmt.Println("Invalid choiche, Please try agian.")
		}
	}
}
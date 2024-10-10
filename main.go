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

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type Priority int

const FILENAME = "db.json"

// Subtask respresents a smaller task within a min task
type Subtasks struct {
	ID         int    `json:"id"`
	Text       string `json:"text"`
	IsComplete bool   `json:"is_complete"`
}

//

// Task respresent a single to-do item with advanced feature
type Task struct {
	Text          string     `json:"name"`
	ID            int        `json:"id"`
	Priority      Priority   `json:"priority"`
	DueDate       time.Time  `json:"due_date"`
	Category      string     `json:"category"`
	IsComplete    bool       `json:"is_complete"`
	Subtasks      []Subtasks `json:"subtask"`
	RecurringCron string     `json:"recuring_cron"`
	Tags          []string   `json:"tags"`
	Attachments   []string   `json:"attachments"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   time.Time  `json:"completed_at"`
	EstimatedTime int        `json:"estimated_time"` // in minutes
	ActualTime    int        `json:"actual_time"`    // in minutes
	Notes         string     `json:"notes"`
	Collaborators []string   `json:"Collaborators"`
}

// Constants
const (
	Low Priority = iota
	Medium
	High
)

type TodoList struct {
	Tasks    []Task `json:"tasks"`
	Password []byte `json:"password"`
}

// Global TodoList
var todos TodoList
var cronSchedular *cron.Cron
var isAuthenticated bool

func init() {
	//set up configuration
	viper.SetConfigName("config")
	viper.SetConfigName("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("storage", "json")
	viper.SetDefault("notification", true)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//config file not found, ingore error
			color.Yellow("Cinfig file not found. Using default settings.")
		} else {
			//config file was found but another error was produced
			color.Red("Error reading config file: %s", err)
		}
	}
}

func main() {
	//load todos
	if err := loadTodo(); err != nil {
		color.Red("Could not load todo from file: %s", err)
	}

	// init cron scheduler
	cronSchedular = cron.New()
	// stop cron schedule the program terminates
	defer cronSchedular.Stop()

	//set up recuring tasks
	setupRecurringTasks()

	//start cron scheduler
	cronSchedular.Start()

	//command line flags
	addFlag := flag.String("add", "", "Add task to list")
	deleteFlag := flag.Int("delete", 0, "Delete item from todo")
	listFlag := flag.Bool("list", false, "List all todos")
	completeFalg := flag.Int("complete", 0, "Filter tasks by category")
	filterFlag := flag.String("filter", "", "Filter tasks by category")
	sortFlag := flag.String("sort", "", "Sort tasks by 'priority', 'due' or 'category'")
	interactiveFlag := flag.Bool("interactive", false, "Enter interactive mode")
	exportFlag := flag.String("export", "", "Export tasks to CSV file")
	importFlag := flag.String("import", "", "Import tasks to CSV file")
	setPasswordFlag := flag.Bool("set-password", false, "Set a password for the todo list")
	generateReportFlag := flag.Bool("report", false, "Generate a productivity report")
	addCollaboratorFlag := flag.String("add-collaborator", "", "Add a collaborator to a tast")
	searchFlag := flag.String("search", "", "Search tasks by keyword")

	flag.Parse()

	//Authenticate user if password is set
	if len(todos.Tasks) > 0 && !isAuthenticated && todos.Password != nil {
		authenticate()
	}

	if *setPasswordFlag {
		setPassword()
	} else if *addFlag != "" {
		add(*addFlag)
	} else if *deleteFlag != 0 {
		remove(*deleteFlag)
	} else if *listFlag {
		list(*filterFlag, *sortFlag)
	} else if *completeFalg != 0 {
		markTaskComplete(*completeFalg)
	} else if *interactiveFlag {
		interactiveMode()
	} else if *exportFlag != "" {
		exportTaskToCSV(*exportFlag)
	} else if *importFlag != "" {
		importTaskFromCSV(*importFlag)
	} else if *generateReportFlag {
		generateProductivityReport()
	} else if *addCollaboratorFlag != "" {
		parts := strings.SplitN(*addCollaboratorFlag, ",", 2)
		if len(parts) != 2 {
			color.Red("Invalid format. Use: -add-collaborator=taskID,collaborator")
		} else {
			taskID, _ := strconv.Atoi(parts[0])
			addCollaborator(taskID, parts[1])
		}
	} else if *searchFlag != "" {
		searchTasks(*searchFlag)
	} else {
		color.Cyan("Please provide a valid command. Use -h for help.")
	}
}

// setup
func setPassword() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter new password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	//hash the password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password: ", err)
		return
	}

	//update db
	todos.Password = hashPassword
	save()
	fmt.Println("Password set successfully.")
}

// authenticate user
func authenticate() {
	reader := bufio.NewReader(os.Stdin)

	//get user input
	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	//decrypt
	err := bcrypt.CompareHashAndPassword(todos.Password, []byte(password))
	if err != nil {
		fmt.Println("Authentication field")
		os.Exit(1)
	}

	isAuthenticated = true
	fmt.Println("Authentication successfully")
}

// set up recuring task
func setupRecurringTasks() {
	for _, task := range todos.Tasks {
		if task.RecurringCron != "" {
			cronSchedular.AddFunc(task.RecurringCron, func() {
				fmt.Printf("Recuring task due: %s\n", task.Text)
			})
		}
	}
}

// load
func loadTodo() error {
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
	return json.Unmarshal(data, &todos)
}

// add collaborator to a task
func addCollaborator(id int, collaborator string) {
	for i, task := range todos.Tasks {
		if task.ID == id {
			todos.Tasks[i].Collaborators = append(todos.Tasks[i].Collaborators, collaborator)
			//update db
			save()
			color.Green("Collaborator %s added to task %d", collaborator, id)
			return
		}
	}

	//tast not found
	color.Red("Task with id: %d not found.", id)
}

// helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(item) {
			return true
		}
	}
	return false
}

// search task by keyword
func searchTasks(keyword string) {
	var results []Task
	for _, task := range todos.Tasks {
		if strings.Contains(strings.ToLower(task.Text), strings.ToLower(keyword)) ||
			strings.Contains(strings.ToLower(task.Category), strings.ToLower(keyword)) ||
			contains(task.Tags, keyword) {
			results = append(results, task)
		}
	}

	if len(results) == 0 {
		color.Yellow("No tasks found matching the keyword: %s", keyword)
		return
	}

	color.Cyan("Search results for keyword: %s", keyword)
	displayTasks(results)
}

//export tast to CSV
func exportTaskToCSV(filename string) {
	file, err := os.Create(filename)

	if err != nil {
		fmt.Println("Error creating file: ", err)
		return
	}

	defer file.Close()

	//write to file
	writer := bufio.NewWriter(file)
	// make sure all data have been saved
	defer writer.Flush()

	//write CSV header
	fmt.Fprintln(writer, "ID, Text, Priority, DueDate, Category, IsComplete, Tags")

	//write task data
	for _, task := range todos.Tasks {
		fmt.Fprintf(writer, "%d,%s,%d,%s,%s,%t,%s\n",
			task.ID, task.Text, task.Priority,
			task.DueDate.Format("2006-01-02"), task.Category,
			task.IsComplete, strings.Join(task.Tags, "|"))
	}

	fmt.Printf("Tasks exported to %s\n", filename)
}

// import from csv
func importTaskFromCSV(filename string) {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip the header

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) != 7 {
			fmt.Println("Invalid CSV format")
			continue
		}

		//destructure field slice
		id, _ := strconv.Atoi(fields[0]) // convert id to string
		priority, _ := strconv.Atoi(fields[2])
		dueDate, _ := time.Parse("2006-01-01", fields[3])
		isComplete, _ := strconv.ParseBool(fields[5])
		tags := strings.Split(fields[6], "|")

		task := Task{
			ID:         id,
			Text:       fields[1],
			Priority:   Priority(priority), //casting priority
			DueDate:    dueDate,
			Category:   fields[4],
			IsComplete: isComplete,
			Tags:       tags,
		}
		//update task
		todos.Tasks = append(todos.Tasks, task)
	}

	//save to db
	save()
	fmt.Print("Task imported from %s\n", filename)
}

// save
func save() error {
	//marsh todo
	//Convert todo to json object
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}
	//return err or nil
	return ioutil.WriteFile(FILENAME, data, 0644)
}

// add
func add(text string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter priority (0: Low, 1: Medium, 2: High): ")
	priorityString, _ := reader.ReadString('\n')
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
		ID:         newID,
		Text:       text,
		Priority:   Priority(priority),
		DueDate:    dueDate,
		Category:   category,
		IsComplete: false,
	})

	//save
	if err := save(); err != nil {
		fmt.Println("Error saving tasks: ", err)
		return
	}

	fmt.Printf("Task added: %s\n", text)
}

// add sub task
func addSubtask(taskId int, subtaskText string) {
	for i, task := range todos.Tasks {
		if task.ID == taskId {
			//generate id subtask
			newSubtaskID := 1
			if len(task.Subtasks) > 0 {
				newSubtaskID = task.Subtasks[len(task.Subtasks)-1].ID + 1
			}
			//append subtask to parent task
			todos.Tasks[i].Subtasks = append(todos.Tasks[i].Subtasks, Subtasks{
				ID:         newSubtaskID,
				Text:       subtaskText,
				IsComplete: false,
			})

			//save task
			save()
			fmt.Printf("Subtask added to task %id: %s\n", newSubtaskID, subtaskText)
			//break from loop
			return
		}
	}
	fmt.Printf("Task with %d id not found", taskId)
}

// remove
func remove(id int) {
	for i, task := range todos.Tasks {
		if task.ID == id {
			//create new slice with the deleted task
			todos.Tasks = append(todos.Tasks[:i], todos.Tasks[i+1:]...)
			//save
			if err := save(); err != nil {
				fmt.Println("Error occured while saving: ", err)
				return
			}
			fmt.Printf("Task removed: %d\n", id)
			return
		}

	}
	fmt.Printf("Task not found: %d\n", id)
}

// display task
func displayTasks(tasks []Task) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description", "Priority", "Due Date", "Category", "Status"})

	for _, task := range tasks {
		status := "Pending"
		if task.IsComplete {
			status = "Completed"
		}

		priority := ""
		switch task.Priority {
		case Low:
			priority = color.GreenString("Low")
		case Medium:
			priority = color.YellowString("Meduim")
		case High:
			priority = color.RedString("High")
		}

		dueDate := task.DueDate.Format("2006-01-02")
		if task.DueDate.Before(time.Now()) && !task.IsComplete {
			dueDate = color.RedString(dueDate + " (Overdue)")
		}

		table.Append([]string{
			strconv.Itoa(task.ID),
			task.Text,
			priority,
			dueDate,
			task.Category,
			status,
		})
	}
	table.Render() //show table
}

// filter task
func filterTask(tasks []Task, filterCategory string) []Task {

	if filterCategory == "" {
		return tasks
	}
	var results []Task

	for _, task := range tasks {
		if task.Category == filterCategory {
			results = append(results, task)
		}
	}

	return results

}

// sort taks
func sortTask(tasks []Task, sortBy string) {
	switch sortBy {
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Priority > tasks[j].Priority
		})
	case "due":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].DueDate.Before(tasks[j].DueDate)
		})
	case "category":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Category < tasks[j].Category
		})
	}

}

// list
func list(filterCategory, sortBy string) {
	if len(todos.Tasks) == 0 {
		fmt.Println("No task")
		return
	}
	filteredTask := filterTask(todos.Tasks, filterCategory)
	sortTask(filteredTask, sortBy)
	displayTasks(filteredTask)
	// filteredTasks := todos.Tasks

	// if filterCategory != "" {
	// 	var tasks []Task
	// 	for _, task := range todos.Tasks {
	// 		if task.Category == filterCategory {
	// 			tasks = append(tasks, task)
	// 		}
	// 	}
	// 	filteredTasks = tasks
	// }

	// switch sortBy {
	// case "priority":
	// 	sort.Slice(filteredTasks, func(i, j int) bool {
	// 		return filteredTasks[i].Priority > filteredTasks[j].Priority
	// 	})
	// case "due":
	// 	sort.Slice(filteredTasks, func(i, j int) bool {
	// 		return filteredTasks[i].DueDate.Before(filteredTasks[j].DueDate)
	// 	})
	// case "category":
	// 	sort.Slice(filteredTasks, func(i, j int) bool {
	// 		return filteredTasks[i].Category < filteredTasks[j].Category
	// 	})
	// }

	// for _, task := range filteredTasks {
	// 	status := " "
	// 	if task.IsComplete {
	// 		status = "✓"
	// 	}

	// 	fmt.Printf("[%s] %d: %s (Priority: %d, Due: %s, Category: %s)\n",
	// 		status, task.ID, task.Text, task.Priority, task.DueDate.Format("2006-01-02"), task.Category)
	// 	//print sub task
	// 	for _, subtask := range task.Subtasks {
	// 		status = " "
	// 		if task.IsComplete {
	// 			status = "✓"
	// 		}
	// 		fmt.Printf(`%5s [%s] %d : %s\n`, "", status, subtask.ID, subtask.Text)
	// 	}
	// }

}

// mark as complete
func markTaskComplete(id int) {
	for i, task := range todos.Tasks {
		if task.ID == id {
			todos.Tasks[i].IsComplete = true
			if err := save(); err != nil {
				fmt.Println("Error saving tasks: ", err)
				return
			}
			fmt.Println("Task marked as complete: ", id)
			return
		}
	}
}

// attachment to task
func addAttachment(taskID int, filename string) {
	for i, task := range todos.Tasks {
		if task.ID == taskID {
			todos.Tasks[i].Attachments = append(todos.Tasks[i].Attachments, filename)
			//update db
			save()
			fmt.Printf("Attachment added to task %d: %s\n", taskID, filename)
			return
		}
	}
	fmt.Printf("Task not found: %d\n", taskID)
}

// generate report
func generateProductivityReport() {
	completedTasks := 0
	totalEstimatedTime := 0
	totalActualTime := 0

	for _, task := range todos.Tasks {
		if task.IsComplete {
			completedTasks++
			totalActualTime += task.ActualTime
			totalEstimatedTime += task.EstimatedTime

		}
	}

	//construct table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Value"})
	table.Append([]string{"Completed Tasks", fmt.Sprintf("%d", completedTasks)})
	table.Append([]string{"Total Estimated Time", fmt.Sprintf("%d", totalEstimatedTime)})
	table.Append([]string{"Total Actual Time", fmt.Sprintf("%d", totalActualTime)})

	//
	if totalEstimatedTime > 0 {
		efficiency := float64(totalEstimatedTime) / float64(totalActualTime) * 100
		table.Append([]string{"Efficiency", fmt.Sprintf("%.2f%%", efficiency)})
	}
	fmt.Println("Called")
	//show table
	table.Render()
}

// interactive mode
func interactiveMode() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nInteractive Mode: ")
		fmt.Println("1. Add task")
		fmt.Println("2. Remove task")
		fmt.Println("3. List tasks")
		fmt.Println("4. Mark task as complete")
		fmt.Println("5: Add subtask")
		fmt.Println("6: Export tasks to CSV")
		fmt.Println("7: Import tasks from CSV")
		fmt.Println("8: Add Attachment A File")
		fmt.Println("9: Generate Productivity report")
		fmt.Println("10: Add collaborator to task")
		fmt.Println("11: Search tasks")
		fmt.Println("12: Exit")
		color.Green("Enter your choice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			color.Yellow("Enter task description: ")
			text, _ := reader.ReadString('\n')
			add(strings.TrimSpace(text))
		case "2":
			color.Yellow("Enter task ID to remove: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			remove(id)
		case "3":
			color.Yellow("Enter filter category (or leave empty): ")
			filter, _ := reader.ReadString('\n')
			filter = strings.TrimSpace(filter)
			color.Yellow("Enter sort method (priority/due/category, or leave emtpy): ")
			sort, _ := reader.ReadString('\n')
			sort = strings.TrimSpace(sort)
			list(filter, sort)
		case "4":
			color.Yellow("Enter task ID to mark as complete: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			markTaskComplete(id)
		case "5":
			color.Yellow("Enter task ID to ask subtask: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			color.Yellow("Enter subtask description: ")
			text, _ := reader.ReadString('\n')
			addSubtask(id, strings.TrimSpace(text))
		case "6":
			color.Yellow("Enter filename to export tasks: ")
			filename, _ := reader.ReadString('\n')
			exportTaskToCSV(strings.TrimSpace(filename))
		case "7":
			color.Yellow("Enter filename to import tasks:")
			filename, _ := reader.ReadString('\n')
			importTaskFromCSV(strings.TrimSpace(filename))
		case "8":
			color.Yellow("Enter task ID: ")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			color.Yellow("Enter attachment filename: ")
			filename, _ := reader.ReadString('\n')
			addAttachment(id, strings.TrimSpace(filename))
		case "9":
			generateProductivityReport()
		case "10":
			color.Yellow("Exiting interactive mode.")
			idStr, _ := reader.ReadString('\n')
			id, _ := strconv.Atoi(strings.TrimSpace(idStr))
			color.Yellow("Enter collaborator name: ")
			collaborator, _ := reader.ReadString('\n')
			addCollaborator(id, strings.TrimSpace(collaborator))
		case "11":
			color.Yellow("Enter search keyword: ")
			keyword, _ := reader.ReadString('\n')
			searchTasks(strings.TrimSpace(keyword))
		case "12":
			fmt.Println("Exiting interactive mode.")
			return
		default:
			color.Red("Invalid choiche, Please try agian.")
		}
	}
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const FILENAME = "db.json"

type Task struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

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

	flag.Parse()

	if *addFlag != "" {
		add(&todos, *addFlag)
	} else if *deleteFlag != 0 {
		remove(&todos, *deleteFlag)
	} else if *listFlag {
		list(todos)
	} else {
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
	// Generating new id
	newID := 1
	if len(todos.Tasks) > 0 {
		//Get the last task id
		newID = todos.Tasks[len(todos.Tasks)-1].ID + 1
	}

	//append new task to slice
	todos.Tasks = append(todos.Tasks, Task{ID: newID, Name: text})

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
func list(todos TodoList) {
	if len(todos.Tasks) == 0 {
		fmt.Println("No task")
		return
	}

	for _, task := range todos.Tasks {
		fmt.Printf("%d. %s", task.ID, task.Name)
	}
}


# Go Command-Line To-Do List Application

This repository contains two versions of a command-line to-do list application written in Go: a basic version and an advanced version. Both applications allow users to manage their tasks through the command line, with the advanced version offering additional features and complexity.




## Table of Contents

    1. Basic Version
    2. Advanced Version
    3. Installation
    4. Usage
    5. Feature Comparison


## Authors

- [@efrimpong](https://www.efrimpong.com)


## Basic Version

The basic version of the to-do list application provides fundamental task management features:
- Add tasks
- Remove tasks
- List all tasks
- Save and load tasks from a JSON file

## The advanced version builds upon the basic version, adding more complex features:

- Task priorities (Low, Medium, High)
- Due dates for tasks
- Categories for tasks
- Marking tasks as complete
- Filtering and sorting options
- Interactive mode for easier task management

## Installation

To use either version of the application:

  1. Ensure you have Go installed on your system. You can download it from golang.org.
  2. Clone this repository or download the desired version of the `main.go` file.
  3. Navigate to the directory containing the `main.go` file in your terminal.
  4. Compile the application:
    `go build main.go`
  5. This will create an executable file named `main` (or `main.exe` on Windows).

# Usage

## Basic Version

Run the application with one of the following commands:
 - Add a task: `./main -add "Task description`
 - Remove a task: `./main -remove [task_id]`
 - List all tasks: `./main -list`

## Advanced Version

The advanced version supports all commands from the basic version, plus:
  - Mark a task as complete: `/main -complete [task_id]`
  - Filter tasks by category: `./main -list -filter [category]`
  - Sort tasks: `./main -list -sort [priority|due|category]`
  - Enter interactive mode: `./main -interactive`

In interactive mode, follow the on-screen prompts to manage your tasks.

## Feature Comparison

| Feature |  Basic Version  | Advanced Version |
|:-----|:--------:|:------:|
| Add tasks   | ✓ | ✓ |
| Remove tasks   |  ✓  | ✓ |
| List tasks   | ✓ | ✓ |
| JSON file storage  | ✗ | ✓ |
| Due dates   | ✗ | ✓ |
| Categories   | ✗ | ✓ |
| Filter and sort   | ✗ | ✓ |
| Interative mode   | ✗ | ✓ |

Choose the version that best suits your needs. The basic version is great for learning Go basics and simple task management, while the advanced version offers a more feature-rich experience for complex task organization.

Feel free to modify and extend either version to fit your specific requirements!
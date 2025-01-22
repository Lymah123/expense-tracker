# Expense Tracker

This is an expense finance tracker that helps to track expenses and generate monthly reports.

## Prerequisites

- Go 1.23.4 or later
- GCC (for CGO support)
- MSYS2 (for Windows users)

## Installation

1. **Clone the repository**:
   ```sh
   git clone https://github.com/yourusername/expense-tracker.git
   cd expense-tracker
```

2. **Install dependencies**:

`go mod tidy`

3. **Build the application**:

`go build -o expense-tracker`

## Usage

CLI Interface

1. Set environment variables(Winddows):
```
$env:CGO_ENABLED = "1"
$env:Path += ";C:\msys64\mingw64\bin"
```

2. Run the application:
`./expense-tracker`

3. Interact with the CLI:
 - Choose an option from tthe menu:
    1. Add an expense
    2. View expenses
    3. Generate monthly report
    4. Exit

## Example

Expense Tracker

Choose an option:
1. Add expense
2. View expense
3. Generate monthly report
4. Exit

Follow the prompts to add expenses, vew expenses, generate reports, and exit the application.


# Web Interface

1. Run the web server:

`go run main.go`

2. Access the web interface:
  - Open your web browser and go to `http://localhost:8080`.

3. Interact with the web interface:
  - Use the web interface to add expenses, view expenses, and generate monthly reports.

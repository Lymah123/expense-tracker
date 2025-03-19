# Expense Tracker

A comprehensive expense tracking application built with Go, featuring secure authentication, budget management, and detailed financial reporting.

## Features

- Interactive dashboard for expense overview
- Add and manage expenses with categories
- Budget tracking with email alerts
- Digital receipt storage
- Customizable financial reports
- Multi-currency support
- Two-Factor Authentication
- Email notifications

## Prerequisites

- Go 1.6+
- SQLite3
- SendGrid API key
- Firebase credentials

## Installation

1. Clone and enter the repository:

```
git clone https://github.com/yourusername/expense-tracker.git
cd expense-tracker
```

2. Install dependencies
```
go mod download
```

3. Create `.env` file
```
SENDGRID_API_KEY=your_sendgrid_api_key
FIREBASE_CREDENTIALS=path_to_firebase_credentials.json
```

4. Build and run
```
go build
./expense-tracker
```

## Project Structure

```
expense-tracker/
|__ apihandlers/  # API
|__ config/ # Configuration
|__ data/ # Database
|__ firebase/ # Firebase integration
|__ handlers/ # HTTP handlers
|__ middleare/ # Middleware
|__ models/ # Data models
|__ reports/ # Report generation
|__ static/ # Static assets
|__ templates/ # HTML templates
|__ utils/ # Utilities
```

## Usage

1. Start the server: `./expense-tracker`
2. Access at `http://localhost:8080`
3. Register/login to begin tracking expenses

## Key Features

### Dashboard
- Expense summaries
- Budget progress
- Recent transactions

### Expense Management

- Add expenses with categories
- Upload receipts
- View history
- Export data

### Budget Tracking
- Ste monthly budgets
- Email alerts
- Category tracking

### Reports
- Custom report generation
- Multiple export formats
- Spending trends
- Category analysis

## Security

- Two Factor Authentication
- Secure password handling
- Session management
- Role-based access
- Input validation

## License

MIT License - See [LICENSE](LICENSE) for details.

## Support

For support, please open an issue in the GutHub repository.


# Go Banking Service

This is a **simple banking service** implemented in Go that simulates basic banking operations, including:

- **Account creation** with support for multiple currencies.
- **Deposits**, **withdrawals**, and **fund transfers**.
- **Role-based user permissions** (Customer, Banker, Teller, Exchange Manager).
- **Currency exchange** with configurable exchange rates.
- **Backup funds** feature, allowing withdrawals from multiple accounts if one has insufficient balance.
- **Thread-safe operations** using `Mutex` and `RWMutex` to ensure concurrency safety.

---

## **Features**

- **Users and Roles:**
  - **Customer**: Manages their accounts.
  - **Banker**: Creates users and accounts.
  - **Teller**: Performs deposits and withdrawals for customers.
  - **Exchange Manager**: Manages currency exchange and sets exchange rates.

- **Account Operations:**
  - Create multiple accounts per user with different currencies.
  - Deposit and withdraw funds.
  - Transfer funds between accounts with the same currency.

- **Currency Exchange:**
  - Exchange between different currencies using stored exchange rates.

- **Backup Funds:**
  - If enabled, withdrawals can use multiple accounts to cover the required amount.

- **Concurrency Safety:**
  - Thread-safe operations using `sync.Mutex` and `sync.RWMutex` to prevent race conditions.

---

## **Setup Instructions**

1. **Prerequisites:**
   - [Go](https://golang.org/dl/) (version 1.20+)

2. **Clone repo:**

   ```bash
   git clone git@github.com:archix/simple-bank-service.git
   ```

3. **Run the Tests:**

   ```bash
   go test -v
   ```

---

## **Usage**

### **Creating a User**
```go
bank := NewBankService()
bank.CreateUser(1, Customer, true) // Create a Customer with backup fund usage enabled
```

### **Creating an Account**
```go
accID, err := bank.CreateAccount(1, 1000, USD) // Create a USD account with an initial deposit of 1000
if err != nil {
    fmt.Println("Error:", err)
}
```

### **Depositing Funds**
```go
bank.Deposit(1, accID, 500) // Deposit 500 USD into the account
```

### **Withdrawing Funds**
```go
err := bank.Withdraw(1, accID, 200) // Withdraw 200 USD from the account
if err != nil {
    fmt.Println("Error:", err)
}
```

### **Transferring Funds**
```go
bank.Transfer(acc1ID, acc2ID, 100) // Transfer 100 USD from one account to another
```

### **Setting Exchange Rates**
```go
bank.SetExchangeRate(USD, EUR, 0.85) // Set exchange rate from USD to EUR
```

### **Exchanging Currency**
```go
err := bank.ExchangeCurrency(1, acc1ID, acc2ID, 100) // Exchange 100 USD to EUR
if err != nil {
    fmt.Println("Error:", err)
}
```

---

## **Concurrency Example**

### **Concurrent Deposits**
```go
var wg sync.WaitGroup
numDeposits := 10
for i := 0; i < numDeposits; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        bank.Deposit(1, accID, 100)
    }()
}
wg.Wait()
```

---

## **Testing**

Run the following command to execute all tests, including **concurrency tests**:

```bash
go test -v
```

Sample Output:
```
=== RUN   TestConcurrentDeposits
--- PASS: TestConcurrentDeposits (0.01s)
=== RUN   TestConcurrentWithdrawals
--- PASS: TestConcurrentWithdrawals (0.01s)
=== RUN   TestConcurrentTransfers
--- PASS: TestConcurrentTransfers (0.01s)
PASS
ok  	bankservice	0.03s
```

---

## **Project Structure**

```
bankservice/
│
├── service.go        # Core banking service logic
├── service_test.go   # Tests for the banking service
├── go.mod            # Go module file
├── LICENSE           # License details
└── README.md         # Project documentation
```

---

## **Future Improvements**

- **API Layer**: Expose the service via REST or gRPC endpoints.
- **Database Integration**: Store users and accounts in a persistent database (e.g., PostgreSQL).
- **Authentication**: Add user authentication and authorization.
- **Improved Currency Exchange**: Integrate real-time exchange rates using an external API.

---

## **Contributing**

Feel free to open issues or submit pull requests if you'd like to contribute!

---

## **Author**

- **Igor Dakic** – Developer of the Go Banking Service

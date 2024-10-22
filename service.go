package main

import (
	"errors"
	"fmt"
	"sync"
)

// Predefined errors for handling failures.
var (
	ErrNegativeDeposit      = errors.New("initial deposit cannot be negative")
	ErrAccountNotExist      = errors.New("account does not exist")
	ErrInvalidAmount        = errors.New("amount must be positive")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrUnauthorizedAccess   = errors.New("unauthorized access to account")
	ErrCurrencyMismatch     = errors.New("currency mismatch between accounts")
	ErrExchangeRateNotFound = errors.New("exchange rate not found")
)

// Supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	GBP = "GBP"
)

// User roles
const (
	Customer        = "customer"
	Banker          = "banker"
	Teller          = "teller"
	ExchangeManager = "exchange_manager"
)

// User represents a bank user with multiple accounts and optional backup fund usage.
type User struct {
	ID             int
	Role           string
	Accounts       []int // List of account IDs belonging to the user
	UseBackupFunds bool  // If true, withdraw from other accounts when needed
}

// Account stores balance and currency information.
type Account struct {
	balance  float64
	currency string
	mutex    sync.RWMutex
	ownerID  int // User ID of the account owner
}

// BankService manages users, accounts, and currency exchange rates.
type BankService struct {
	accounts      map[int]*Account
	users         map[int]*User
	exchangeRates map[string]float64 // Store exchange rates (e.g., "USD:EUR" -> 0.85)
	nextAccountID int
	mutex         sync.Mutex
}

// NewBankService initializes a new BankService instance.
func NewBankService() *BankService {
	return &BankService{
		accounts:      make(map[int]*Account),
		users:         make(map[int]*User),
		exchangeRates: make(map[string]float64),
	}
}

// CreateUser creates a new user with a specific role and backup fund usage setting.
func (b *BankService) CreateUser(userID int, role string, useBackupFunds bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.users[userID] = &User{
		ID:             userID,
		Role:           role,
		UseBackupFunds: useBackupFunds,
	}
	fmt.Printf("Created user %d with role %s\n", userID, role)
}

// CreateAccount creates an account for a user with an initial deposit and currency.
func (b *BankService) CreateAccount(userID int, initialDeposit float64, currency string) (int, error) {
	if initialDeposit < 0 {
		return 0, ErrNegativeDeposit
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	accountID := b.nextAccountID
	b.accounts[accountID] = &Account{
		balance:  initialDeposit,
		currency: currency,
		ownerID:  userID,
	}
	b.nextAccountID++

	b.users[userID].Accounts = append(b.users[userID].Accounts, accountID)
	fmt.Printf("Created account %d for user %d with %s %.2f\n", accountID, userID, currency, initialDeposit)
	return accountID, nil
}

// CheckPermissions verifies if the user can access the account.
func (b *BankService) CheckPermissions(userID, accountID int) error {
	account, exists := b.accounts[accountID]
	if !exists {
		return ErrAccountNotExist
	}

	user := b.users[userID]
	if user.Role == Banker || account.ownerID == userID {
		return nil // Access granted
	}
	return ErrUnauthorizedAccess
}

// GetBalance retrieves the balance and currency of an account.
func (b *BankService) GetBalance(userID, accountID int) (float64, string, error) {
	if err := b.CheckPermissions(userID, accountID); err != nil {
		return 0, "", err
	}

	account := b.accounts[accountID]
	account.mutex.RLock()
	defer account.mutex.RUnlock()

	return account.balance, account.currency, nil
}

// Deposit adds funds to the specified account.
func (b *BankService) Deposit(userID, accountID int, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if err := b.CheckPermissions(userID, accountID); err != nil {
		return err
	}

	account := b.accounts[accountID]
	account.mutex.Lock()
	defer account.mutex.Unlock()

	account.balance += amount
	fmt.Printf("User %d deposited %.2f to account %d\n", userID, amount, accountID)
	return nil
}

// Withdraw tries to withdraw from the specified account, with optional backup funds usage.
func (b *BankService) Withdraw(userID, accountID int, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if err := b.CheckPermissions(userID, accountID); err != nil {
		return err
	}

	account := b.accounts[accountID]
	account.mutex.Lock()
	defer account.mutex.Unlock()

	if account.balance >= amount {
		account.balance -= amount
		fmt.Printf("User %d withdrew %.2f from account %d\n", userID, amount, accountID)
		return nil
	}

	// Try backup funds if allowed.
	user := b.users[userID]
	if user.UseBackupFunds {
		remaining := amount - account.balance
		account.balance = 0
		fmt.Printf("User %d withdrew %.2f from primary account %d, remaining %.2f\n", userID, amount-account.balance, accountID, remaining)
		return b.withdrawFromOtherAccounts(userID, accountID, remaining)
	}

	return ErrInsufficientBalance
}

// withdrawFromOtherAccounts withdraws the remaining amount from backup accounts.
func (b *BankService) withdrawFromOtherAccounts(userID, excludeAccountID int, amount float64) error {
	for _, accID := range b.users[userID].Accounts {
		if accID == excludeAccountID {
			continue // Skip the original account
		}

		account := b.accounts[accID]
		account.mutex.Lock()
		if account.balance >= amount {
			account.balance -= amount
			account.mutex.Unlock()
			fmt.Printf("Withdrew %.2f from backup account %d\n", amount, accID)
			return nil
		}
		amount -= account.balance
		account.balance = 0
		account.mutex.Unlock()
		fmt.Printf("Used all funds from backup account %d, remaining %.2f\n", accID, amount)
	}

	return ErrInsufficientBalance
}

// Transfer transfers funds between two accounts with the same currency.
func (b *BankService) Transfer(fromID, toID int, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	fromAccount, err := b.getAccount(fromID)
	if err != nil {
		return err
	}

	toAccount, err := b.getAccount(toID)
	if err != nil {
		return err
	}

	if fromAccount.currency != toAccount.currency {
		return ErrCurrencyMismatch
	}

	fromAccount.mutex.Lock()
	defer fromAccount.mutex.Unlock()

	if fromAccount.balance < amount {
		return ErrInsufficientBalance
	}

	toAccount.mutex.Lock()
	defer toAccount.mutex.Unlock()

	fromAccount.balance -= amount
	toAccount.balance += amount
	fmt.Printf("Transferred %.2f from account %d to account %d\n", amount, fromID, toID)
	return nil
}

// SetExchangeRate sets the exchange rate between two currencies.
func (b *BankService) SetExchangeRate(from, to string, rate float64) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	key := from + ":" + to
	b.exchangeRates[key] = rate
	fmt.Printf("Set exchange rate %s -> %s: %.2f\n", from, to, rate)
}

// ExchangeCurrency exchanges an amount from one currency to another.
func (b *BankService) ExchangeCurrency(userID, fromID, toID int, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if err := b.CheckPermissions(userID, fromID); err != nil {
		return err
	}
	if err := b.CheckPermissions(userID, toID); err != nil {
		return err
	}

	fromAccount := b.accounts[fromID]
	toAccount := b.accounts[toID]

	key := fromAccount.currency + ":" + toAccount.currency
	rate, exists := b.exchangeRates[key]
	if !exists {
		return ErrExchangeRateNotFound
	}

	fromAccount.mutex.Lock()
	defer fromAccount.mutex.Unlock()

	if fromAccount.balance < amount {
		return ErrInsufficientBalance
	}

	toAccount.mutex.Lock()
	defer toAccount.mutex.Unlock()

	fromAccount.balance -= amount
	toAccount.balance += amount * rate
	fmt.Printf("Exchanged %.2f %s to %.2f %s\n", amount, fromAccount.currency, amount*rate, toAccount.currency)
	return nil
}

// getAccount retrieves an account by its ID.
func (b *BankService) getAccount(accountID int) (*Account, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	account, exists := b.accounts[accountID]
	if !exists {
		return nil, ErrAccountNotExist
	}
	return account, nil
}

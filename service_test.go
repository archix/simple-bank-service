package main

import (
	"errors"
	"sync"
	"testing"
)

// TestCreateUser ensures that users are created with correct roles and settings.
func TestCreateUser(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, true)

	user, exists := bank.users[1]
	if !exists {
		t.Fatalf("expected user 1 to exist")
	}

	if user.Role != Customer {
		t.Errorf("expected user role to be %s, got %s", Customer, user.Role)
	}
	if !user.UseBackupFunds {
		t.Errorf("expected UseBackupFunds to be true")
	}
}

// TestCreateAccount ensures accounts are created with correct balances and currencies.
func TestCreateAccount(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)

	accID, err := bank.CreateAccount(1, 1000, USD)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance, currency, _ := bank.GetBalance(1, accID)
	if balance != 1000 || currency != USD {
		t.Errorf("expected balance 1000 and currency USD, got %.2f and %s", balance, currency)
	}
}

// TestCreateAccountNegativeDeposit checks that creating an account with negative deposit fails.
func TestCreateAccountNegativeDeposit(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)

	_, err := bank.CreateAccount(1, -100, USD)
	if !errors.Is(err, ErrNegativeDeposit) {
		t.Fatalf("expected ErrNegativeDeposit, got %v", err)
	}
}

// TestDeposit ensures that deposits increase account balance.
func TestDeposit(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)
	accID, _ := bank.CreateAccount(1, 500, USD)

	err := bank.Deposit(1, accID, 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance, _, _ := bank.GetBalance(1, accID)
	if balance != 700 {
		t.Errorf("expected balance 700, got %.2f", balance)
	}
}

// TestWithdraw ensures that withdrawals reduce the account balance.
func TestWithdraw(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)
	accID, _ := bank.CreateAccount(1, 500, USD)

	err := bank.Withdraw(1, accID, 300)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance, _, _ := bank.GetBalance(1, accID)
	if balance != 200 {
		t.Errorf("expected balance 200, got %.2f", balance)
	}
}

func TestWithdrawWithBackupFunds(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, true)

	acc1, _ := bank.CreateAccount(1, 100, USD) // Primary account
	acc2, _ := bank.CreateAccount(1, 50, USD)  // Backup account 1
	acc3, _ := bank.CreateAccount(1, 100, USD) // Backup account 2

	// Attempt to withdraw 200, using backup funds if needed
	err := bank.Withdraw(1, acc1, 200)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance1, _, _ := bank.GetBalance(1, acc1)
	balance2, _, _ := bank.GetBalance(1, acc2)
	balance3, _, _ := bank.GetBalance(1, acc3)

	if balance1 != 0 || balance2 != 0 || balance3 != 50 {
		t.Errorf("expected balances to be 0, 0, and 50; got %.2f, %.2f, and %.2f", balance1, balance2, balance3)
	}
}

// TestTransfer ensures successful transfer between accounts with the same currency.
func TestTransfer(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)

	acc1, _ := bank.CreateAccount(1, 1000, USD)
	acc2, _ := bank.CreateAccount(1, 500, USD)

	err := bank.Transfer(acc1, acc2, 300)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance1, _, _ := bank.GetBalance(1, acc1)
	balance2, _, _ := bank.GetBalance(1, acc2)

	if balance1 != 700 || balance2 != 800 {
		t.Errorf("expected balances to be 700 and 800, got %.2f and %.2f", balance1, balance2)
	}
}

// TestTransferCurrencyMismatch ensures transfer fails between accounts with different currencies.
func TestTransferCurrencyMismatch(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)

	acc1, _ := bank.CreateAccount(1, 1000, USD)
	acc2, _ := bank.CreateAccount(1, 500, EUR)

	err := bank.Transfer(acc1, acc2, 100)
	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
}

// TestSetAndUseExchangeRate ensures currency exchange works with valid rates.
func TestSetAndUseExchangeRate(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, ExchangeManager, false)

	acc1, _ := bank.CreateAccount(1, 1000, USD)
	acc2, _ := bank.CreateAccount(1, 0, EUR)

	bank.SetExchangeRate(USD, EUR, 0.85)

	err := bank.ExchangeCurrency(1, acc1, acc2, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	balance1, _, _ := bank.GetBalance(1, acc1)
	balance2, _, _ := bank.GetBalance(1, acc2)

	if balance1 != 900 || balance2 != 85 {
		t.Errorf("expected balances to be 900 and 85, got %.2f and %.2f", balance1, balance2)
	}
}

// TestExchangeRateNotFound ensures exchange fails if the rate is not found.
func TestExchangeRateNotFound(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, ExchangeManager, false)

	acc1, _ := bank.CreateAccount(1, 1000, USD)
	acc2, _ := bank.CreateAccount(1, 0, GBP)

	err := bank.ExchangeCurrency(1, acc1, acc2, 100)
	if !errors.Is(err, ErrExchangeRateNotFound) {
		t.Fatalf("expected ErrExchangeRateNotFound, got %v", err)
	}
}

// TestUnauthorizedAccess ensures unauthorized users can't access accounts.
func TestUnauthorizedAccess(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)
	bank.CreateUser(2, Teller, false)

	accID, _ := bank.CreateAccount(1, 500, USD)

	err := bank.Withdraw(2, accID, 100) // Teller shouldn't have access
	if !errors.Is(err, ErrUnauthorizedAccess) {
		t.Fatalf("expected ErrUnauthorizedAccess, got %v", err)
	}
}

// TestConcurrentDeposits ensures concurrent deposits work without race conditions.
func TestConcurrentDeposits(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)
	accID, _ := bank.CreateAccount(1, 0, USD)

	var wg sync.WaitGroup
	numDeposits := 10
	amountPerDeposit := 100.0

	// Start multiple goroutines to deposit concurrently.
	for i := 0; i < numDeposits; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := bank.Deposit(1, accID, amountPerDeposit); err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		}()
	}

	wg.Wait() // Wait for all deposits to complete.

	balance, _, _ := bank.GetBalance(1, accID)
	expectedBalance := float64(numDeposits) * amountPerDeposit
	if balance != expectedBalance {
		t.Errorf("expected balance %.2f, got %.2f", expectedBalance, balance)
	}
}

// TestConcurrentWithdrawals tests concurrent withdrawals with backup funds.
func TestConcurrentWithdrawals(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, true)

	// Create multiple accounts to serve as backup.
	acc1, _ := bank.CreateAccount(1, 300, USD) // Primary account
	acc2, _ := bank.CreateAccount(1, 200, USD) // Backup account

	var wg sync.WaitGroup
	numWithdrawals := 5
	amountPerWithdrawal := 100.0

	// Start multiple goroutines to withdraw concurrently.
	for i := 0; i < numWithdrawals; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bank.Withdraw(1, acc1, amountPerWithdrawal)
			if err != nil && !errors.Is(err, ErrInsufficientBalance) {
				t.Errorf("expected no error or insufficient balance, got %v", err)
			}
		}()
	}

	wg.Wait() // Wait for all withdrawals to complete.

	// Verify the remaining balances after withdrawals.
	balance1, _, _ := bank.GetBalance(1, acc1)
	balance2, _, _ := bank.GetBalance(1, acc2)

	if balance1 != 0 || balance2 != 0 {
		t.Errorf("expected balances to be 0 and 0, got %.2f and %.2f", balance1, balance2)
	}
}

// TestConcurrentTransfers tests multiple concurrent transfers between accounts.
func TestConcurrentTransfers(t *testing.T) {
	bank := NewBankService()
	bank.CreateUser(1, Customer, false)

	acc1, _ := bank.CreateAccount(1, 1000, USD)
	acc2, _ := bank.CreateAccount(1, 500, USD)

	var wg sync.WaitGroup
	numTransfers := 10
	amountPerTransfer := 50.0

	// Start multiple goroutines to transfer concurrently.
	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := bank.Transfer(acc1, acc2, amountPerTransfer)
			if err != nil && !errors.Is(err, ErrInsufficientBalance) {
				t.Errorf("expected no error or insufficient balance, got %v", err)
			}
		}()
	}

	wg.Wait() // Wait for all transfers to complete.

	// Verify the final balances.
	balance1, _, _ := bank.GetBalance(1, acc1)
	balance2, _, _ := bank.GetBalance(1, acc2)

	expectedBalance1 := 1000 - float64(numTransfers)*amountPerTransfer
	expectedBalance2 := 500 + float64(numTransfers)*amountPerTransfer

	if balance1 != expectedBalance1 || balance2 != expectedBalance2 {
		t.Errorf("expected balances to be %.2f and %.2f, got %.2f and %.2f",
			expectedBalance1, expectedBalance2, balance1, balance2)
	}
}

package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"log"
	"sync"
	"time"
)

type DashboardService struct {
	Users                UserSvc
	Account              AccountSvc
	DashboardGetHandler  fiber.Handler
	DashboardPostHandler fiber.Handler
}

type UserSvc struct{}

func (svc *UserSvc) GetStats(context.Context, string) (UserStats, error) {
	return UserStats{}, nil
}

func (svc *UserSvc) GetTransactions(context.Context, string) <-chan Transaction {
	return make(chan Transaction)
}

type AccountSvc struct{}

func (svc *AccountSvc) GetStats(context.Context, string) <-chan AccountStats {
	return make(chan AccountStats)
}

func (svc *AccountSvc) GetTransactions(context.Context, string) <-chan Transaction {
	return make(chan Transaction)
}

type DashboardData struct {
	UserData         UserStats
	AccountData      AccountStats
	LastTransactions []Transaction
}

type DashboardParams struct{}

type UserStats struct{}

type AccountStats struct{}

type Transaction struct{}

func (svc *DashboardService) GetDashboardData(ctx context.Context, userID string) DashboardData {
	result := DashboardData{}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		result.UserData, err = svc.Users.GetStats(ctx, userID)
		if err != nil {
			log.Println(err)
		}
	}()

	acctCh := make(chan AccountStats)
	go func() {
		defer close(acctCh)
		newCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		resultCh := svc.Account.GetStats(newCtx, userID)
		select {
		case data := <-resultCh:
			acctCh <- data
		case <-newCtx.Done():
		}
	}()

	transactionWg := sync.WaitGroup{}
	transactionWg.Add(1)
	transactionCh := make(chan Transaction)
	go func() {
		defer transactionWg.Done()
		for t := range svc.Users.GetTransactions(ctx, userID) {
			transactionCh <- t
		}
	}()
	go func() {
		transactionWg.Wait()
		close(transactionCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for record := range transactionCh {
			result.LastTransactions = append(result.LastTransactions, record)
		}
	}()

	wg.Wait()
	result.AccountData = <-acctCh
	return result
}

func (svc *DashboardService) SetDashboardConfig(ctx context.Context, userID string, params DashboardParams) error {
	return nil
}

func NewDashboardService() *DashboardService {
	svc := DashboardService{}
	svc.DashboardGetHandler = func(c *fiber.Ctx) error {
		dashboard := svc.GetDashboardData(c.Context(), c.Params("userID"))
		return c.JSON(dashboard)
	}
	svc.DashboardPostHandler = func(c *fiber.Ctx) error {
		var params DashboardParams
		err := c.BodyParser(params)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		err = svc.SetDashboardConfig(c.Context(), c.Params("userID"), params)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.SendStatus(fiber.StatusOK)
	}
	return &svc
}

func Limit(maxSize int, next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Request().Header.ContentLength() > maxSize {
			return c.SendStatus(fiber.StatusRequestEntityTooLarge)
		}

		return next(c)
	}
}

func Auth(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authKey := c.Request().Header.Peek("Authorization")
		if authKey == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		//err := authorize(authKey)
		//if err != nil {
		//	return c.SendStatus(fiber.StatusUnauthorized)
		//}

		return next(c)
	}
}

func ConcurrencyLimiter(sem chan struct{}, next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sem <- struct{}{}
		defer func() { <-sem }()

		return next(c)
	}
}

func main() {
	app := fiber.New()
	svc := NewDashboardService()

	semaphore := make(chan struct{}, 10)
	app.Get("/dashboard", ConcurrencyLimiter(semaphore, Limit(1024, Auth(svc.DashboardGetHandler))))
	app.Post("/dashboard", ConcurrencyLimiter(semaphore, Limit(1024, Auth(svc.DashboardPostHandler))))
	app.Listen("localhost:10001")
}

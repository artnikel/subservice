package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/artnikel/subservice/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

var (
	db       *pgxpool.Pool
	repo     *SubscriptionRepository
	pool     *dockertest.Pool
	resource *dockertest.Resource
	ctx      = context.Background()
	sub      *models.Subscription
)

func TestMain(m *testing.M) {
	var err error

	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %v", err)
	}

	resource, err = pool.Run("postgres", "15", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=testdb",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %v", err)
	}

	err = pool.Retry(func() error {
		connStr := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))
		db, err = pgxpool.New(context.Background(), connStr)
		if err != nil {
			return err
		}
		return db.Ping(context.Background())
	})
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	createTables()

	repo = NewSubscriptionRepository(db)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %v", err)
	}

	os.Exit(code)
}

func createTables() {
	_, err := db.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		service_name TEXT NOT NULL,
		price INTEGER NOT NULL,
		user_id UUID NOT NULL,
		start_date TEXT NOT NULL,
		end_date TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT now(),
		updated_at TIMESTAMP NOT NULL DEFAULT now()
	);
	`)
	if err != nil {
		log.Fatalf("Could not create tables: %v", err)
	}
}

func cleanTables(t *testing.T) {
	_, err := db.Exec(context.Background(), "DELETE FROM subscriptions")
	assert.NoError(t, err)
}

func createTestSubscription(t *testing.T) *models.Subscription {
	endDate := "2024-12"
	sub := &models.Subscription{
		ServiceName: "Spotify",
		Price:       999,
		UserID:      uuid.New(),
		StartDate:   "2024-01",
		EndDate:     &endDate,
	}
	err := repo.Create(ctx, sub)
	assert.NoError(t, err)
	return sub
}

func TestCreateSubscription(t *testing.T) {
	defer cleanTables(t)
	endDate := "2024-12"
	sub = &models.Subscription{
		ServiceName: "Spotify",
		Price:       999,
		UserID:      uuid.New(),
		StartDate:   "2024-01",
		EndDate:     &endDate,
	}
	err := repo.Create(ctx, sub)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, sub.ID)
}

func TestGetSubscriptionByID(t *testing.T) {
	defer cleanTables(t)

	sub := createTestSubscription(t)

	got, err := repo.GetByID(ctx, sub.ID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, sub.ServiceName, got.ServiceName)
}

func TestUpdateSubscription(t *testing.T) {
	defer cleanTables(t)

	sub := createTestSubscription(t)

	newName := "Netflix"
	newPrice := 1299
	updates := &models.SubscriptionUpdates{
		ServiceName: &newName,
		Price:       &newPrice,
	}
	updated, err := repo.Update(ctx, sub.ID, updates)
	assert.NoError(t, err)
	assert.Equal(t, newName, updated.ServiceName)
	assert.Equal(t, newPrice, updated.Price)
}

func TestListSubscriptions(t *testing.T) {
	defer cleanTables(t)

	sub := createTestSubscription(t)

	list, total, err := repo.List(ctx, &sub.UserID, nil, 1, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(list), 1)
}

func TestGetCostSummary(t *testing.T) {
	defer cleanTables(t)

	sub := createTestSubscription(t)
	userIDstring := sub.UserID.String()
	cost, err := repo.GetCostSummary(ctx, &userIDstring, nil, "2024-01", "2024-12")
	assert.NoError(t, err)
	assert.Equal(t, sub.Price, cost)
}

func TestDeleteSubscription(t *testing.T) {
	defer cleanTables(t)

	sub := createTestSubscription(t)

	err := repo.Delete(ctx, sub.ID)
	assert.NoError(t, err)

	got, err := repo.GetByID(ctx, sub.ID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

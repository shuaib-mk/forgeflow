//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/forgeflow/forgeflow/internal/database"
	"github.com/forgeflow/forgeflow/internal/migrations"
)

func TestMigrationsAreIdempotent(t *testing.T){databaseURL:=os.Getenv("FORGEFLOW_DATABASE_URL");if databaseURL==""{t.Skip("FORGEFLOW_DATABASE_URL is not set")};db,err:=database.Open(context.Background(),databaseURL);if err!=nil{t.Fatal(err)};defer db.Close();for attempt:=0;attempt<2;attempt++{if err:=migrations.Up(context.Background(),db.Pool,"../../deployments/migrations");err!=nil{t.Fatalf("migration attempt %d: %v",attempt+1,err)}}}


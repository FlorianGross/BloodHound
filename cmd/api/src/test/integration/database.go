// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/cache"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func OpenDatabase(t *testing.T) database.Database {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		t.Fatalf("Failed loading integration test config: %v", err)
	} else if db, err := database.OpenDatabase(cfg.Database.PostgreSQLConnectionString()); err != nil {
		t.Fatalf("Failed to open database: %v", err)
	} else {
		return database.NewBloodhoundDB(db, auth.NewIdentityResolver())
	}

	return nil
}

func OpenCache(t *testing.T) cache.Cache {
	if cache, err := cache.NewCache(cache.Config{MaxSize: 200}); err != nil {
		t.Fatalf("Failed creating cache: %e", err)
	} else {
		return cache
	}
	return cache.Cache{}
}

func SetupDB(t *testing.T) database.Database {
	dbInst := OpenDatabase(t)
	if err := Prepare(context.Background(), dbInst); err != nil {
		t.Fatalf("Error preparing DB: %v", err)
	}
	return dbInst
}

func Prepare(ctx context.Context, db database.Database) error {
	if err := db.Wipe(ctx); err != nil {
		return fmt.Errorf("failed to clear database: %v", err)
	} else if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

func bootstrapGraphDb(ctx context.Context, cfg config.Configuration) error {
	if graphDB, err := bootstrap.ConnectGraph(ctx, cfg); err != nil {
		return fmt.Errorf("failed to connect graph database: %v", err)
	} else {
		defer graphDB.Close(ctx)
		return bootstrap.MigrateGraph(ctx, graphDB, schema.DefaultGraphSchema())
	}
}

func SetupTestMigrator(sources ...migration.Source) (*gorm.DB, *migration.Migrator, error) {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to load integration test config: %w", err)
	} else if db, err := gorm.Open(postgres.Open(cfg.Database.PostgreSQLConnectionString())); err != nil {
		return nil, nil, fmt.Errorf("failed to open postgres connection: %w", err)
	} else if err = wipeGormDB(db); err != nil {
		return nil, nil, fmt.Errorf("failed to wipe database: %w", err)
	} else if err := bootstrapGraphDb(context.Background(), cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to bootstrap graph db database: %w", err)
	} else {
		return db, &migration.Migrator{
			Sources: sources,
			DB:      db,
		}, nil
	}
}

func wipeGormDB(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		sql := `
				do $$ declare
					r record;
				begin
					for r in (select tablename from pg_tables where schemaname = 'public') loop
						execute 'drop table if exists ' || quote_ident(r.tablename) || ' cascade';
					end loop;
				end $$;
			`
		return tx.Exec(sql).Error
	})
}

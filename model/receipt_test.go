// model/receipt_test.go

package model

import (
	"context"
	"encoding/json"
	"testing"

	"fmt"
	"os"
	"reflect"
	"time"

	"rcpt-proc-challenge-ans/config"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Initialize zap.Logger before running tests
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	config.Log = logger

	// Set up the test database
    testDB, err := setupTestDatabase()
    if err != nil {
        fmt.Printf("Failed to set up test database: %v\n", err)
		config.Log.Error("Failed to set up test database: %v\n", zap.Error(err))
        os.Exit(1)
    }

	// Set the global DB variable to use in tests
    config.DB = testDB

	// Truncate tables before running tests
	if err := truncateTables(config.DB); err != nil {
		config.Log.Error("Failed to truncate tables", zap.Error(err))
		fmt.Printf("Failed to truncate tables: %v\n", err)
	}

    // Run the tests
    code := m.Run()

	// Teardown
    tearDownTestDatabase(testDB)

    // Exit with the test result code
    os.Exit(code)
}

func TestValidateReceipt(t *testing.T) {
	for _, testCase := range ValidateReceiptTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var receipt Receipt
			err := json.Unmarshal([]byte(testCase.JsonData), &receipt)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			err = receipt.ValidateReceipt()
			isValid := err == nil

			if isValid != testCase.IsValid {
				t.Errorf("Test case %s failed. Expected isValid to be %v, got %v", testCase.Name, testCase.IsValid, isValid)
			}
		})
	}
}


// Test case for UUID generation with GenerateID
func TestGenerateID(t *testing.T) {
	receipt := Receipt{}
	receipt.GenerateID()

	// Create a zero UUID instance
	zeroUUID := uuid.UUID{}

	if receipt.ID == zeroUUID {
		t.Error("expected a non-nil UUID, got nil UUID")
	}
}


func TestCalculatePoints(t *testing.T) {
	for _, testCase := range CalculatePointsTestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var receipt Receipt
			err := json.Unmarshal([]byte(testCase.JsonData), &receipt)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Clean the item descriptions before calculating points
            receipt.CleanItemShortDescriptions()

			receipt.CalculatePoints()

			if receipt.Points != testCase.ExpectedPoints {
				t.Errorf("Test case %s failed. Expected %d points, got %d", testCase.Name, testCase.ExpectedPoints, receipt.Points)
			}
		})
	}
}


func TestCleanItemShortDescription(t *testing.T) {
    testCases := []struct {
        name     string
        input    []Item
        expected []Item
    }{
        {
            name: "Clean descriptions",
            input: []Item{
                {ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  "},
                {ShortDescription: "Mountain Dew 12PK"},
            },
            expected: []Item{
                {ShortDescription: "Klarbrunn 12-PK 12 FL OZ"},
                {ShortDescription: "Mountain Dew 12PK"},
            },
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            receipt := Receipt{Items: tc.input}
            receipt.CleanItemShortDescriptions()
            
            if !reflect.DeepEqual(receipt.Items, tc.expected) {
                t.Errorf("Expected %v, got %v", tc.expected, receipt.Items)
            }
        })
    }
}

// Calculation Tests:
func TestCalculatePointsFromRetailerAlphaNumChar(t *testing.T) {
    testCases := []struct {
        retailer string
        expected uint
    }{
        {"Target", 6},
        {"123", 3},
        {"Walmart!", 7},
        {"", 0},
    }

    for _, tc := range testCases {
        result := calculatePointsFromRetailerAlphaNumChar(tc.retailer)
        if result != tc.expected {
            t.Errorf("For retailer %s, expected %d, got %d", tc.retailer, tc.expected, result)
        }
    }
}

func TestCalculatePointsFromTotal(t *testing.T) {
    testCases := []struct {
        total    string
        expected uint
    }{
        {"10.00", 75},
        {"10.25", 25},
        {"10.50", 25},
        {"10.75", 25},
        {"10.99", 0},
    }

    for _, tc := range testCases {
        result := calculatePointsFromTotal(tc.total)
        if result != tc.expected {
            t.Errorf("For total %s, expected %d, got %d", tc.total, tc.expected, result)
        }
    }
}

/*
	Test SKU Methods:
*/
func TestSKUParsing(t *testing.T) {
    for _, tc := range SKUParsingTestCases {
        t.Run(tc.Name, func(t *testing.T) {
            var sku SKU
            err := sku.ParseSKU(tc.Input)
            if tc.ExpectError {
                if err == nil {
                    t.Errorf("Expected error for input %s, got nil", tc.Input)
                }
            } else {
                if err != nil {
                    t.Fatalf("Failed to parse SKU %s: %v", tc.Input, err)
                }
                if !reflect.DeepEqual(sku, tc.Expected) {
                    t.Errorf("Expected %+v, got %+v", tc.Expected, sku)
                }
            }
        })
    }
}

/*
	Test Endpoint Methods: 
	GET, 
	POST
*/
func TestDatabaseOperations(t *testing.T) {
	// Truncate tables before this test
	if err := truncateTables(config.DB); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

    t.Run("TestAddReceipt", func(t *testing.T) {
        receipt := createTestReceipt()

        // Log the receipt before adding
        receiptJSON, _ := json.MarshalIndent(receipt, "", "  ")
        t.Logf("Adding receipt:\n%s", string(receiptJSON))

        err := AddReceipt(config.DB, receipt)
        if err != nil {
            t.Errorf("Failed to add receipt: %v", err)
        }
    })

    t.Run("TestGetReceiptByID", func(t *testing.T) {
        receipt := createTestReceipt()
        err := AddReceipt(config.DB, receipt)
        if err != nil {
            t.Fatalf("Failed to add receipt: %v", err)
        }

        fetchedReceipt, err := GetReceiptByID(config.DB, receipt.ID)
        if err != nil {
            t.Errorf("Failed to get receipt by ID: %v", err)
        }

		// Log the fetched receipt
        fetchedJSON, _ := json.MarshalIndent(fetchedReceipt, "", "  ")
        t.Logf("Fetched receipt:\n%s", string(fetchedJSON))

        if !reflect.DeepEqual(receipt, fetchedReceipt) {
            t.Errorf("Fetched receipt does not match original")

			// Log both receipts for comparison
            originalJSON, _ := json.MarshalIndent(receipt, "", "  ")
            t.Logf("Original receipt:\n%s", string(originalJSON))
            t.Logf("Fetched receipt:\n%s", string(fetchedJSON))
        }
    })

    t.Run("TestGetAllReceipts", func(t *testing.T) {
        // Clear existing receipts
        _, err := config.DB.Exec(context.Background(), "DELETE FROM items; DELETE FROM receipts;")
        if err != nil {
            t.Fatalf("Failed to clear existing receipts: %v", err)
        }

        // Add multiple receipts
        for i := 0; i < 3; i++ {
            receipt := createTestReceipt()
            err := AddReceipt(config.DB, receipt)
            if err != nil {
                t.Fatalf("Failed to add receipt: %v", err)
            }
        }

        receipts, err := GetAllReceipts(config.DB)
        if err != nil {
            t.Errorf("Failed to get all receipts: %v", err)
        }

		// Log all fetched receipts
        for i, r := range receipts {
            receiptJSON, _ := json.MarshalIndent(r, "", "  ")
            t.Logf("Receipt %d:\n%s", i+1, string(receiptJSON))
        }

        if len(receipts) != 3 {
            t.Errorf("Expected 3 receipts, got %d", len(receipts))
        }
    })
}

func createTestReceipt() *Receipt {
	receiptID := config.GenerateUUID()

	// Retrieve the current count of items
    itemsCount, err := GetItemsCount(config.DB)
    if err != nil {
        fmt.Printf("Failed to get items count: %v", err)
    }

    return &Receipt{
        ID:           receiptID,
        Retailer:     "Test Store",
        PurchaseDate: time.Now().Format("2006-01-02"),
        PurchaseTime: time.Now().Format("15:04"),
        Items: []Item{
            {
				ID: uint((itemsCount + 1)),
                SKU: SKU{
                    Prefix:           "TST",
                    ProductCategory:  "GROC",
                    Manufacturer:     "TESTBRAND",
                    ProductLine:      "PROD",
                    Attributes:       map[string]string{"SIZE": "LRG"},
                    UniqueIdentifier: "12345",
                },
                ShortDescription: "Test Product Large",
                Quantity:         1,
                PricePaid:        "10.00",
				ReceiptID: 		  receiptID,
            },
            {
				ID: uint((itemsCount + 2)),
                SKU: SKU{
                    Prefix:           "TST",
                    ProductCategory:  "ELEC",
                    Manufacturer:     "TESTTECH",
                    ProductLine:      "GADGET",
                    Attributes:       map[string]string{"COLOR": "RED"},
                    UniqueIdentifier: "67890",
                },
                ShortDescription: "Test Gadget Red",
                Quantity:         2,
                PricePaid:        "25.00",
				ReceiptID: 		  receiptID,
            },
        },
        Total:  "60.00",
        Points: 0, // Points will be calculated later
    }
}

func setupTestDatabase() (*pgxpool.Pool, error) {
    // Read database connection details from environment variables
    // You might want to set these in your CI/CD pipeline or local development environment
    dbConfig := getDBConfig()

    // Construct the database connection string
    dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", 
        dbConfig["user"], dbConfig["password"], dbConfig["host"], dbConfig["port"], dbConfig["dbname"])

    // Create a database connection pool
    config, err := pgxpool.ParseConfig(dbURL)
    if err != nil {
        return nil, fmt.Errorf("error parsing database config: %v", err)
    }

    // Set some connection pool settings
    config.MaxConns = 5
    config.MaxConnLifetime = time.Hour

    // Create the connection pool
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("error creating database pool: %v", err)
    }

    // Test the connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("error connecting to the database: %v", err)
    }

    // Optional: Set up the database schema
    if err := setupTestSchema(pool); err != nil {
        return nil, fmt.Errorf("error setting up test schema: %v", err)
    }

    return pool, nil
}

func setupTestSchema(pool *pgxpool.Pool) error {
    // Create necessary tables for testing
    _, err := pool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS receipts (
            id UUID PRIMARY KEY,
            retailer VARCHAR(255) NOT NULL,
            purchase_date DATE NOT NULL,
            purchase_time TIME NOT NULL,
            total DECIMAL(10, 2) NOT NULL,
            points INTEGER NOT NULL
        );

        CREATE TABLE IF NOT EXISTS skus (
            unique_identifier VARCHAR(255) PRIMARY KEY,
            prefix VARCHAR(50) NOT NULL,
            product_category VARCHAR(50) NOT NULL,
            manufacturer VARCHAR(50) NOT NULL,
            product_line VARCHAR(50) NOT NULL,
            attributes JSONB
        );

        CREATE TABLE IF NOT EXISTS items (
            id SERIAL PRIMARY KEY,
            short_description VARCHAR(255) NOT NULL,
            quantity INTEGER NOT NULL,
            price_paid DECIMAL(10, 2) NOT NULL,
            receipt_id UUID REFERENCES receipts(id),
            sku_id VARCHAR(255) REFERENCES skus(unique_identifier)
        );
    `)

    return err
}

func tearDownTestDatabase(pool *pgxpool.Pool) {
    if pool != nil {
        pool.Close()

        // Wait for up to 5 seconds for the pool to close
        for i := 0; i < 5; i++ {
            ctx, cancel := context.WithTimeout(context.Background(), time.Second)
            err := pool.Ping(ctx)
            cancel()

            if err != nil {
                fmt.Println("Database pool successfully closed")
                return
            }

            time.Sleep(time.Second)
        }

        fmt.Println("Warning: Database pool did not close within 5 seconds")
    }
}

func truncateTables(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		TRUNCATE TABLE items, receipts RESTART IDENTITY CASCADE;
	`)
	return err
}

func getDBConfig() map[string]string {
    config := map[string]string{
        "host":     "localhost",
        "port":     "5435", // Default PostgreSQL port
        "user":     "",
        "password": "",
        "dbname":   "",
    }

    // Override with environment variables if they exist
    if host := os.Getenv("TEST_DB_HOST"); host != "" {
        config["host"] = host
    }
    if port := os.Getenv("TEST_DB_PORT"); port != "" {
        config["port"] = port
    }
    if user := os.Getenv("TEST_DB_USER"); user != "" {
        config["user"] = user
    }
    if password := os.Getenv("TEST_DB_PASS"); password != "" {
        config["password"] = password
    }
    if dbname := os.Getenv("TEST_DB_NAME"); dbname != "" {
        config["dbname"] = dbname
    }

    return config
}
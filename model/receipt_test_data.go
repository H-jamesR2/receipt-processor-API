// model/receipt_test_data.go

package model

var ValidateReceiptTestCases = []struct {
    Name            string
    JsonData        string
    IsValid         bool
}{
    {
        Name: "Valid Receipt",
        JsonData: `{
            "retailer": "Target",
            "purchaseDate": "2022-01-01",
            "purchaseTime": "13:01",
            "total": "6.49",
            "items": [
                {
                    "shortDescription": "Mountain Dew 12PK",
                    "quantity": 1,
                    "pricePaid": "6.49",
                    "sku": "TGT-BVRG-MTNDEW-SODA-SIZE-12PK-00001"
                }
            ]
        }`,
        IsValid: true,
    },
    {   
        Name: "Empty Retailer",
        JsonData: `{
            "retailer": "",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "13:13",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "14.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Date",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "13/13/2023",
            "purchaseTime": "13:13",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "14.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Time",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "24:01",
            "total": "14.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "14.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Mismatched Total vs. ItemsTotal",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "11.00",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "6.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Invalid Total",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "InvalidTotal",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "6.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid: false,
    },
    {
        Name: "Missing Item Price",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "10.00",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid:       false,
    },
        {
        Name: "Missing Item Description",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "10.00",
            "items": [
                {
                    "quantity": 10,
                    "pricePaid": "10.00",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid:       false,
    },
    {
        Name: "Negative Item Price",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "6.25",
            "items": [
                {
                    "shortDescription": "Pepsi - 12-oz",
                    "quantity": 10,
                    "pricePaid": "-6.25",
                    "sku": "WMT-BVRG-PEPSI-SODA-SIZE-12OZ-00001"
                }
            ]
        }`,
        IsValid:       false,
    },
    {
        Name: "Missing Items List",
        JsonData: `{
            "retailer": "Walmart",
            "purchaseDate": "2022-01-02",
            "purchaseTime": "05:00",
            "total": "6.25",
            "items": []
        }`,
        IsValid:       false,
    },
}

var CalculatePointsTestCases = []struct {
    Name           string
    JsonData       string
    ExpectedPoints uint
}{
    {
        Name: "Target Receipt",
        JsonData: `{
            "retailer": "Target",
            "purchaseDate": "2022-01-01",
            "purchaseTime": "13:01",
            "total": "35.35",
            "items": [
                {
                    "shortDescription": "Mountain Dew 12PK",
                    "quantity": 1,
                    "pricePaid": "6.49",
                    "sku": "TGT-BVRG-MTNDEW-SODA-SIZE-12PK-00001"
                },
                {
                    "shortDescription": "Emils Cheese Pizza",
                    "quantity": 1,
                    "pricePaid": "12.25",
                    "sku": "TGT-FOOD-EMILS-PIZZA-TYPE-CHEESE-00002"
                },
                {
                    "shortDescription": "Knorr Creamy Chicken",
                    "quantity": 1,
                    "pricePaid": "1.26",
                    "sku": "TGT-FOOD-KNORR-SOUP-FLVR-CHICKEN-00003"
                },
                {
                    "shortDescription": "Doritos Nacho Cheese",
                    "quantity": 1,
                    "pricePaid": "3.35",
                    "sku": "TGT-SNCK-DORITOS-CHIPS-FLVR-NACHO-00004"
                },
                {
                    "shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
                    "quantity": 1,
                    "pricePaid": "12.00",
                    "sku": "TGT-BVRG-KLARBRUNN-WATER-SIZE-12PK-00005"
                }
            ]
        }`,
        ExpectedPoints: 28, 
    },
    // Add more test cases here
}

var SKUParsingTestCases = []struct {
    Name         string
    Input        string
    Expected     SKU
    ExpectError  bool
}{
    {
        Name:  "Valid SKU",
        Input: "WMT-GROC-NESTLE-CHOC-WEIGHT-100G-67890",
        Expected: SKU{
            Prefix:           "WMT",
            ProductCategory:  "GROC",
            Manufacturer:     "NESTLE",
            ProductLine:      "CHOC",
            Attributes:       map[string]string{"WEIGHT": "100G"},
            UniqueIdentifier: "67890",
        },
        ExpectError: false,
    },
    {
        Name:        "Malformed SKU",
        Input:       "TGT-GROC",
        ExpectError: true,
    },
    // Add more SKU parsing test cases...
}
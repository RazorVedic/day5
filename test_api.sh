#!/bin/bash

# API Testing Script for Retailer Service
# Make sure the service is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1"

echo "🚀 Testing Retailer Service API"
echo "================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to make API calls and show responses
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "→ $method $endpoint"
    
    if [ -n "$data" ]; then
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" \
            -w "\nHTTP_CODE:%{http_code}")
    else
        response=$(curl -s -X $method "$BASE_URL$endpoint" \
            -w "\nHTTP_CODE:%{http_code}")
    fi
    
    # Extract HTTP code
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    json_response=$(echo "$response" | sed '/HTTP_CODE:/d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Success ($http_code)${NC}"
        echo "$json_response" | jq '.' 2>/dev/null || echo "$json_response"
    else
        echo -e "${RED}✗ Error ($http_code)${NC}"
        echo "$json_response" | jq '.' 2>/dev/null || echo "$json_response"
    fi
}

# Check if service is running
echo "🔍 Checking if service is running..."
if curl -s "$BASE_URL/../health" > /dev/null; then
    echo -e "${GREEN}✓ Service is running${NC}"
else
    echo -e "${RED}✗ Service is not running. Please start it first with 'make docker-run' or 'make run'${NC}"
    exit 1
fi

# # 1. Add Products (Retailer)
# echo -e "\n${YELLOW}=== RETAILER: PRODUCT MANAGEMENT ===${NC}"

# test_endpoint "POST" "/product" '{
#   "product_name": "iPhone 15 Pro",
#   "price": 999.99,
#   "quantity": 25
# }' "Add iPhone 15 Pro"

# test_endpoint "POST" "/product" '{
#   "product_name": "MacBook Air M3",
#   "price": 1199.99,
#   "quantity": 10
# }' "Add MacBook Air M3"

# test_endpoint "POST" "/product" '{
#   "product_name": "AirPods Pro",
#   "price": 249.99,
#   "quantity": 50
# }' "Add AirPods Pro"

# 2. View Products (Customers & Retailer)
test_endpoint "GET" "/products" "" "View all products"

# Store product IDs for later use (extract from previous response)
PRODUCT_ID1="PROD70516"  # Note: These are example IDs, actual IDs will be auto-generated
PRODUCT_ID2="PROD29905"
PRODUCT_ID3="PROD52432"

echo -e "\n${YELLOW}📝 Note: Using example product IDs. In real testing, extract IDs from previous responses.${NC}"

# 3. Update Product (Retailer)
test_endpoint "PUT" "/product/$PRODUCT_ID1" '{
  "price": 899.99,
  "quantity": 30
}' "Update iPhone price and quantity"

# # 4. Register Customers
# echo -e "\n${YELLOW}=== CUSTOMER MANAGEMENT ===${NC}"

# test_endpoint "POST" "/customer" '{
#   "name": "Alice Johnson",
#   "email": "alice@example.com",
#   "phone": "+1234567890"
# }' "Register Alice"

# test_endpoint "POST" "/customer" '{
#   "name": "Bob Smith",
#   "email": "bob@example.com",
#   "phone": "+1987654321"
# }' "Register Bob"

# 5. View Customers (Retailer)
test_endpoint "GET" "/customers" "" "View all customers"

# Store customer IDs
CUSTOMER_ID1="CUST38290"  # Example IDs
CUSTOMER_ID2="CUST28233"

echo -e "\n${YELLOW}📝 Note: Using example customer IDs. In real testing, extract IDs from previous responses.${NC}"

# 6. Place Orders (Customers)
echo -e "\n${YELLOW}=== ORDER MANAGEMENT ===${NC}"

test_endpoint "POST" "/order" '{
  "customer_id": "'$CUSTOMER_ID1'",
  "product_id": "'$PRODUCT_ID1'",
  "quantity": 1
}' "Alice orders iPhone"

test_endpoint "POST" "/order" '{
  "customer_id": "'$CUSTOMER_ID2'",
  "product_id": "'$PRODUCT_ID3'",
  "quantity": 2
}' "Bob orders AirPods"

# 7. Test Cooldown (should fail)
echo -e "\n${YELLOW}🕐 Testing cooldown mechanism...${NC}"
test_endpoint "POST" "/order" '{
  "customer_id": "'$CUSTOMER_ID1'",
  "product_id": "'$PRODUCT_ID2'",
  "quantity": 1
}' "Alice tries to order again (should fail due to cooldown)"

# 8. View Order History
echo -e "\n${YELLOW}=== ORDER HISTORY ===${NC}"

test_endpoint "GET" "/orders/customer/$CUSTOMER_ID1" "" "Alice's order history"
test_endpoint "GET" "/orders" "" "All orders (Retailer view)"

# 9. Transaction History & Analytics (Retailer)
echo -e "\n${YELLOW}=== BUSINESS ANALYTICS ===${NC}"

test_endpoint "GET" "/transactions" "" "Transaction history"
test_endpoint "GET" "/transactions/stats" "" "Business statistics dashboard"

# 10. Test filtering
test_endpoint "GET" "/transactions?limit=5&type=order" "" "Filtered transactions (orders only)"

echo -e "\n${GREEN}🎉 API Testing Complete!${NC}"
echo -e "\n${YELLOW}Next Steps:${NC}"
echo "1. Check the actual IDs from the responses above"
echo "2. Replace example IDs in this script with real ones for more accurate testing"
echo "3. Wait 5+ minutes and test the cooldown mechanism again"
echo "4. Try the API documentation for more advanced filtering options"

echo -e "\n${YELLOW}📚 For complete API documentation, see: API_DOCUMENTATION.md${NC}"

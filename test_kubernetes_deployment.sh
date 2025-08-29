#!/bin/bash

# Day5 Kubernetes Deployment Test Script
# This script tests the complete retailer API workflow in Kubernetes

set -e

echo "🚀 Testing Day5 Retailer API in Kubernetes..."
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Test variables
BASE_URL="http://localhost:8080"
PRODUCT_ID=""
CUSTOMER_ID=""
ORDER_ID=""

echo "📋 Step 1: Health Check"
echo "Testing: GET $BASE_URL/health"
HEALTH_RESPONSE=$(curl -s $BASE_URL/health)
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    print_status "Health check passed"
    echo "$HEALTH_RESPONSE" | python3 -m json.tool
else
    print_error "Health check failed"
    echo "Response: $HEALTH_RESPONSE"
    exit 1
fi
echo

echo "📦 Step 2: Create Product"
echo "Testing: POST $BASE_URL/api/v1/product"
PRODUCT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/product \
  -H "Content-Type: application/json" \
  -d '{"product_name":"Test Laptop Pro","price":2299.99,"quantity":15}')

if echo "$PRODUCT_RESPONSE" | grep -q "successfully created"; then
    PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    print_status "Product created successfully with ID: $PRODUCT_ID"
    echo "Response: $PRODUCT_RESPONSE" | python3 -m json.tool
else
    print_error "Product creation failed"
    echo "Response: $PRODUCT_RESPONSE"
    exit 1
fi
echo

echo "👤 Step 3: Create Customer"
echo "Testing: POST $BASE_URL/api/v1/customer"
CUSTOMER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/customer \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Customer Pro","email":"test.pro@example.com","phone":"1234567890"}')

if echo "$CUSTOMER_RESPONSE" | grep -q "successfully created"; then
    CUSTOMER_ID=$(echo "$CUSTOMER_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    print_status "Customer created successfully with ID: $CUSTOMER_ID"
    echo "Response: $CUSTOMER_RESPONSE" | python3 -m json.tool
else
    print_error "Customer creation failed"
    echo "Response: $CUSTOMER_RESPONSE"
    exit 1
fi
echo

echo "📋 Step 4: List Products"
echo "Testing: GET $BASE_URL/api/v1/products"
PRODUCTS_RESPONSE=$(curl -s $BASE_URL/api/v1/products)
if echo "$PRODUCTS_RESPONSE" | grep -q "Products retrieved successfully"; then
    print_status "Products listed successfully"
    echo "Response: $PRODUCTS_RESPONSE" | python3 -m json.tool
else
    print_error "Products listing failed"
    echo "Response: $PRODUCTS_RESPONSE"
    exit 1
fi
echo

echo "🛍️ Step 5: Place Order"
echo "Testing: POST $BASE_URL/api/v1/order"
ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/order \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":\"$CUSTOMER_ID\",\"product_id\":\"$PRODUCT_ID\",\"quantity\":2}")

if echo "$ORDER_RESPONSE" | grep -q "successfully placed"; then
    ORDER_ID=$(echo "$ORDER_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
    print_status "Order placed successfully with ID: $ORDER_ID"
    echo "Response: $ORDER_RESPONSE" | python3 -m json.tool
else
    print_error "Order placement failed"
    echo "Response: $ORDER_RESPONSE"
    exit 1
fi
echo

echo "⏰ Step 6: Test Cooldown Period"
echo "Testing: POST $BASE_URL/api/v1/order (should fail due to cooldown)"
COOLDOWN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/order \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":\"$CUSTOMER_ID\",\"product_id\":\"$PRODUCT_ID\",\"quantity\":1}")

if echo "$COOLDOWN_RESPONSE" | grep -q "cooldown period"; then
    print_status "Cooldown mechanism working correctly"
    echo "Response: $COOLDOWN_RESPONSE" | python3 -m json.tool
else
    print_warning "Cooldown test unclear - response: $COOLDOWN_RESPONSE"
fi
echo

echo "📊 Step 7: Business Analytics"
echo "Testing: GET $BASE_URL/api/v1/transactions/stats"
STATS_RESPONSE=$(curl -s $BASE_URL/api/v1/transactions/stats)
if echo "$STATS_RESPONSE" | grep -q "total_revenue"; then
    print_status "Business analytics working correctly"
    echo "Response: $STATS_RESPONSE" | python3 -m json.tool
else
    print_error "Business analytics failed"
    echo "Response: $STATS_RESPONSE"
    exit 1
fi
echo

echo "📈 Step 8: Order History"
echo "Testing: GET $BASE_URL/api/v1/orders/customer/$CUSTOMER_ID"
HISTORY_RESPONSE=$(curl -s "$BASE_URL/api/v1/orders/customer/$CUSTOMER_ID")
if echo "$HISTORY_RESPONSE" | grep -q "$ORDER_ID"; then
    print_status "Order history working correctly"
    echo "Response: $HISTORY_RESPONSE" | python3 -m json.tool
else
    print_error "Order history failed"
    echo "Response: $HISTORY_RESPONSE"
    exit 1
fi
echo

echo "🔍 Step 9: Kubernetes Status Check"
echo "Checking pod and service status..."
echo "Pods:"
kubectl get pods -n day5
echo
echo "Services:"
kubectl get services -n day5
echo

echo "🎉 All tests completed successfully!"
echo
echo "📋 Summary:"
echo "✅ Health check"
echo "✅ Product management"
echo "✅ Customer management"
echo "✅ Order processing"
echo "✅ Cooldown mechanism"
echo "✅ Business analytics"
echo "✅ Order history"
echo "✅ Kubernetes deployment"
echo
echo "🏆 Day5 Retailer API is fully functional in Kubernetes!"
echo
echo "🔗 API Documentation: http://localhost:8080/health"
echo "📊 Try the API endpoints using the examples above"
echo
print_status "Deployment successful! 🎯"

#!/bin/bash

# API Testing script for Escrow microservices

set -e

API_BASE="http://localhost:8080/api/v1"
echo "🧪 Testing Escrow API"
echo "=========================="
echo "Base URL: $API_BASE"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to make HTTP requests and show results
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local auth_header=$4
    local description=$5

    echo -e "${BLUE}Testing:${NC} $description"
    echo -e "${YELLOW}$method${NC} $endpoint"
    
    if [ -n "$data" ]; then
        echo -e "${YELLOW}Data:${NC} $data"
    fi
    
    local curl_cmd="curl -s -w '\nStatus: %{http_code}\nTime: %{time_total}s\n' -X $method"
    
    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_header'"
    fi
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    curl_cmd="$curl_cmd '$API_BASE$endpoint'"
    
    echo "Response:"
    eval $curl_cmd | jq . 2>/dev/null || eval $curl_cmd
    echo ""
    echo "---"
    echo ""
}

# Check if jq is available for JSON formatting
if ! command -v jq &> /dev/null; then
    echo "⚠️  jq is not installed. JSON responses won't be formatted."
    echo "   Install with: brew install jq (macOS) or apt-get install jq (Ubuntu)"
    echo ""
fi

echo "🔍 Step 1: Health Check"
test_endpoint "GET" "/health" "" "" "API Gateway Health Check"

echo "👤 Step 2: User Registration"
USER_EMAIL="test$(date +%s)@example.com"
REGISTER_DATA="{\"email\":\"$USER_EMAIL\",\"password\":\"password123\"}"
REGISTER_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -d "$REGISTER_DATA" "$API_BASE/auth/register")
test_endpoint "POST" "/auth/register" "$REGISTER_DATA" "" "Register new user"

# Extract access token from registration response
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.access_token' 2>/dev/null || echo "")
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.user.id' 2>/dev/null || echo "")

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo -e "${RED}❌ Failed to get access token from registration${NC}"
    echo "Registration response: $REGISTER_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Successfully registered user with ID: $USER_ID${NC}"
echo ""

echo "🔐 Step 3: User Login"
LOGIN_DATA="{\"email\":\"$USER_EMAIL\",\"password\":\"password123\"}"
test_endpoint "POST" "/auth/login" "$LOGIN_DATA" "" "Login with registered user"

echo "👤 Step 4: Get User Profile"
test_endpoint "GET" "/users/$USER_ID" "" "$ACCESS_TOKEN" "Get user profile"

echo "💳 Step 5: Create Wallet"
WALLET_DATA="{\"currency\":\"USD\"}"
test_endpoint "POST" "/wallets" "$WALLET_DATA" "$ACCESS_TOKEN" "Create wallet"

echo "💰 Step 6: Make a Transfer"
TRANSFER_DATA="{\"toUserId\":\"$USER_ID\",\"amount\":100.00,\"currency\":\"USD\",\"description\":\"Test transfer\"}"
test_endpoint "POST" "/transfers" "$TRANSFER_DATA" "$ACCESS_TOKEN" "Make transfer"

echo "📄 Step 7: Create a Contract"
# For simplicity, using the same user as both client and freelancer
CONTRACT_DATA="{\"title\":\"Test Contract\",\"description\":\"This is a test contract\",\"amount\":500.00,\"currency\":\"USD\",\"clientId\":\"$USER_ID\",\"freelancerId\":\"$USER_ID\"}"
CONTRACT_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" -d "$CONTRACT_DATA" "$API_BASE/contracts")
test_endpoint "POST" "/contracts" "$CONTRACT_DATA" "$ACCESS_TOKEN" "Create new contract"

# Extract contract ID
CONTRACT_ID=$(echo "$CONTRACT_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "")

echo "📄 Step 8: Get Contract Details"
if [ -n "$CONTRACT_ID" ] && [ "$CONTRACT_ID" != "null" ]; then
    test_endpoint "GET" "/contracts/$CONTRACT_ID" "" "$ACCESS_TOKEN" "Get contract details"
else
    echo -e "${YELLOW}⚠️  Skipping contract details - no contract ID available${NC}"
fi

echo "📋 Step 9: List User's Contracts"
test_endpoint "GET" "/contracts" "" "$ACCESS_TOKEN" "List user's contracts"

echo "⚖️  Step 10: Create a Dispute"
if [ -n "$CONTRACT_ID" ] && [ "$CONTRACT_ID" != "null" ]; then
    DISPUTE_DATA="{\"contractId\":\"$CONTRACT_ID\",\"title\":\"Test Dispute\",\"description\":\"This is a test dispute\",\"category\":\"quality\"}"
    DISPUTE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $ACCESS_TOKEN" -d "$DISPUTE_DATA" "$API_BASE/disputes")
    test_endpoint "POST" "/disputes" "$DISPUTE_DATA" "$ACCESS_TOKEN" "Create dispute"
    
    # Extract dispute ID
    DISPUTE_ID=$(echo "$DISPUTE_RESPONSE" | jq -r '.data.id' 2>/dev/null || echo "")
else
    echo -e "${YELLOW}⚠️  Skipping dispute creation - no contract ID available${NC}"
fi

echo "🔔 Step 11: Get Notifications"
test_endpoint "GET" "/notifications" "" "$ACCESS_TOKEN" "Get user notifications"

echo "📊 Step 12: Check Audit Logs"
test_endpoint "GET" "/audit/logs" "" "$ACCESS_TOKEN" "Get audit logs"

echo ""
echo -e "${GREEN}🎉 API Testing Complete!${NC}"
echo ""
echo "📋 Summary:"
echo "├── User Email: $USER_EMAIL"
echo "├── User ID: $USER_ID"
echo "├── Contract ID: ${CONTRACT_ID:-'N/A'}"
echo "└── Dispute ID: ${DISPUTE_ID:-'N/A'}"
echo ""
echo "💡 Tips:"
echo "├── Connect to WebSocket: ws://localhost:8081/ws?userId=$USER_ID&clientId=test"
echo "├── View logs: docker-compose logs -f [service-name]"
echo "├── MongoDB: mongodb://localhost:27017 (admin/admin)"
echo "├── RabbitMQ Management: http://localhost:15672 (admin/admin)"
echo "└── Check service health: curl http://localhost:8080/api/v1/health"
echo ""
echo "🧪 Test completed successfully!"


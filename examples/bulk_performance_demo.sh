#!/bin/bash

# Bulk Operations Performance Demo
# This script demonstrates the performance of prepared statements
# in bulk operations using the gosql CLI tool

set -e

echo "=== Bulk Operations Performance Demo ==="
echo ""

# Clean start
rm -f data.db

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Bulk insert users
echo -e "${BLUE}Test 1: Bulk inserting 100 users${NC}"
cat > /tmp/100_users.json << 'EOF'
[
  {"email":"user0@example.com","name":"User 0","age":20},
  {"email":"user1@example.com","name":"User 1","age":21},
  {"email":"user2@example.com","name":"User 2","age":22},
  {"email":"user3@example.com","name":"User 3","age":23},
  {"email":"user4@example.com","name":"User 4","age":24},
  {"email":"user5@example.com","name":"User 5","age":25},
  {"email":"user6@example.com","name":"User 6","age":26},
  {"email":"user7@example.com","name":"User 7","age":27},
  {"email":"user8@example.com","name":"User 8","age":28},
  {"email":"user9@example.com","name":"User 9","age":29},
  {"email":"user10@example.com","name":"User 10","age":30},
  {"email":"user11@example.com","name":"User 11","age":31},
  {"email":"user12@example.com","name":"User 12","age":32},
  {"email":"user13@example.com","name":"User 13","age":33},
  {"email":"user14@example.com","name":"User 14","age":34},
  {"email":"user15@example.com","name":"User 15","age":35},
  {"email":"user16@example.com","name":"User 16","age":36},
  {"email":"user17@example.com","name":"User 17","age":37},
  {"email":"user18@example.com","name":"User 18","age":38},
  {"email":"user19@example.com","name":"User 19","age":39},
  {"email":"user20@example.com","name":"User 20","age":40},
  {"email":"user21@example.com","name":"User 21","age":41},
  {"email":"user22@example.com","name":"User 22","age":42},
  {"email":"user23@example.com","name":"User 23","age":43},
  {"email":"user24@example.com","name":"User 24","age":44},
  {"email":"user25@example.com","name":"User 25","age":45},
  {"email":"user26@example.com","name":"User 26","age":46},
  {"email":"user27@example.com","name":"User 27","age":47},
  {"email":"user28@example.com","name":"User 28","age":48},
  {"email":"user29@example.com","name":"User 29","age":49},
  {"email":"user30@example.com","name":"User 30","age":50},
  {"email":"user31@example.com","name":"User 31","age":51},
  {"email":"user32@example.com","name":"User 32","age":52},
  {"email":"user33@example.com","name":"User 33","age":53},
  {"email":"user34@example.com","name":"User 34","age":54},
  {"email":"user35@example.com","name":"User 35","age":55},
  {"email":"user36@example.com","name":"User 36","age":56},
  {"email":"user37@example.com","name":"User 37","age":57},
  {"email":"user38@example.com","name":"User 38","age":58},
  {"email":"user39@example.com","name":"User 39","age":59},
  {"email":"user40@example.com","name":"User 40","age":60},
  {"email":"user41@example.com","name":"User 41","age":61},
  {"email":"user42@example.com","name":"User 42","age":62},
  {"email":"user43@example.com","name":"User 43","age":63},
  {"email":"user44@example.com","name":"User 44","age":64},
  {"email":"user45@example.com","name":"User 45","age":65},
  {"email":"user46@example.com","name":"User 46","age":66},
  {"email":"user47@example.com","name":"User 47","age":67},
  {"email":"user48@example.com","name":"User 48","age":68},
  {"email":"user49@example.com","name":"User 49","age":69}
]
EOF

# Add 50 more users to make 100
for i in {50..99}; do
  age=$((20 + i % 50))
  sed -i.bak "$ s/]$/,{\"email\":\"user${i}@example.com\",\"name\":\"User ${i}\",\"age\":${age}}]/" /tmp/100_users.json
done

./bin/gosql user-bulk-upsert -f /tmp/100_users.json
echo ""

# Test 2: Bulk update (upsert)
echo -e "${BLUE}Test 2: Bulk upserting 50 existing users (should UPDATE)${NC}"
cat > /tmp/50_users_update.json << 'EOF'
[
  {"email":"user0@example.com","name":"UPDATED User 0","age":99},
  {"email":"user1@example.com","name":"UPDATED User 1","age":99},
  {"email":"user2@example.com","name":"UPDATED User 2","age":99}
]
EOF

# Add more update records
for i in {3..49}; do
  sed -i.bak "$ s/]$/,{\"email\":\"user${i}@example.com\",\"name\":\"UPDATED User ${i}\",\"age\":99}]/" /tmp/50_users_update.json
done

./bin/gosql user-bulk-upsert -f /tmp/50_users_update.json
echo ""

# Test 3: Bulk insert courses
echo -e "${BLUE}Test 3: Bulk inserting 20 courses${NC}"
cat > /tmp/20_courses.json << 'EOF'
[
  {"slug":"go-basics","title":"Go Programming Basics","price":100},
  {"slug":"python-advanced","title":"Advanced Python","price":250},
  {"slug":"rust-intro","title":"Introduction to Rust","price":150},
  {"slug":"javascript-fullstack","title":"Full Stack JavaScript","price":300},
  {"slug":"typescript-deep","title":"Deep Dive into TypeScript","price":200},
  {"slug":"docker-mastery","title":"Docker Mastery","price":180},
  {"slug":"kubernetes-basics","title":"Kubernetes Fundamentals","price":220},
  {"slug":"aws-cloud","title":"AWS Cloud Practitioner","price":275},
  {"slug":"react-hooks","title":"React Hooks Masterclass","price":190},
  {"slug":"vue-composition","title":"Vue 3 Composition API","price":170},
  {"slug":"svelte-complete","title":"Complete Svelte Course","price":160},
  {"slug":"nodejs-backend","title":"Node.js Backend Development","price":240},
  {"slug":"graphql-api","title":"GraphQL API Design","price":210},
  {"slug":"postgresql-advanced","title":"Advanced PostgreSQL","price":195},
  {"slug":"mongodb-nosql","title":"MongoDB for NoSQL","price":185},
  {"slug":"redis-caching","title":"Redis Caching Strategies","price":155},
  {"slug":"nginx-reverse-proxy","title":"Nginx Configuration","price":145},
  {"slug":"microservices-arch","title":"Microservices Architecture","price":320},
  {"slug":"testing-strategies","title":"Testing Best Practices","price":175},
  {"slug":"ci-cd-pipelines","title":"CI/CD Pipelines","price":230}
]
EOF

./bin/gosql course-bulk-upsert -f /tmp/20_courses.json
echo ""

# Test 4: Mixed upsert (some new, some updates)
echo -e "${BLUE}Test 4: Mixed upsert - 5 updates + 5 new courses${NC}"
cat > /tmp/mixed_courses.json << 'EOF'
[
  {"slug":"go-basics","title":"Go UPDATED","price":999},
  {"slug":"python-advanced","title":"Python UPDATED","price":888},
  {"slug":"new-course-1","title":"Brand New Course 1","price":500},
  {"slug":"new-course-2","title":"Brand New Course 2","price":550},
  {"slug":"new-course-3","title":"Brand New Course 3","price":600}
]
EOF

./bin/gosql course-bulk-upsert -f /tmp/mixed_courses.json
echo ""

# Summary
echo -e "${GREEN}=== Summary ===${NC}"
echo "Prepared statements provide:"
echo "  ✓ Consistent performance across operations"
echo "  ✓ Protection against SQL injection"
echo "  ✓ Efficient batch processing"
echo ""
echo "Check the operation_time and avg_per_record in the output above!"
echo ""

# Cleanup
rm -f /tmp/100_users.json /tmp/100_users.json.bak
rm -f /tmp/50_users_update.json /tmp/50_users_update.json.bak
rm -f /tmp/20_courses.json /tmp/mixed_courses.json

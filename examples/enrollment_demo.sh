#!/bin/bash

# Course Enrollment Demo with Transactions
# This script demonstrates how transactions ensure data integrity

set -e

echo "=== Course Enrollment Transaction Demo ==="
echo ""

# Clean start
rm -f data.db

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Step 1: Create test data
echo -e "${BLUE}Step 1: Creating users and courses${NC}"
./bin/gosql user-add -e "alice@example.com" -n "Alice Smith" -a 25 > /dev/null
./bin/gosql user-add -e "bob@example.com" -n "Bob Jones" -a 30 > /dev/null
./bin/gosql course-add -s "go-basics" -t "Go Programming Basics" -p 100 > /dev/null
./bin/gosql course-add -s "python-advanced" -t "Advanced Python" -p 250 > /dev/null
./bin/gosql course-add -s "rust-intro" -t "Introduction to Rust" -p 150 > /dev/null
echo -e "${GREEN}✓ Created 2 users and 3 courses${NC}"
echo ""

# Step 2: Successful enrollment (transaction commits)
echo -e "${BLUE}Step 2: Enrolling Alice in Go Basics (should succeed)${NC}"
./bin/gosql enrollment-create -u 1 -c 1
echo -e "${GREEN}✓ Transaction committed successfully${NC}"
echo ""

# Step 3: Failed enrollment - non-existent user (transaction rolls back)
echo -e "${BLUE}Step 3: Trying to enroll non-existent user ID 999 (should fail)${NC}"
./bin/gosql enrollment-create -u 999 -c 1 2>&1 || echo -e "${RED}✓ Transaction rolled back - no enrollment created${NC}"
echo ""

# Step 4: Failed enrollment - non-existent course (transaction rolls back)
echo -e "${BLUE}Step 4: Trying to enroll Alice in non-existent course ID 999 (should fail)${NC}"
./bin/gosql enrollment-create -u 1 -c 999 2>&1 || echo -e "${RED}✓ Transaction rolled back - no enrollment created${NC}"
echo ""

# Step 5: Failed enrollment - duplicate (transaction rolls back)
echo -e "${BLUE}Step 5: Trying to enroll Alice in Go Basics again (should fail - already enrolled)${NC}"
./bin/gosql enrollment-create -u 1 -c 1 2>&1 || echo -e "${RED}✓ Transaction rolled back - UNIQUE constraint prevented duplicate${NC}"
echo ""

# Step 6: Multiple successful enrollments
echo -e "${BLUE}Step 6: Enrolling Alice in more courses${NC}"
./bin/gosql enrollment-create -u 1 -c 2 > /dev/null
./bin/gosql enrollment-create -u 1 -c 3 > /dev/null
echo -e "${GREEN}✓ Alice now enrolled in 3 courses${NC}"
echo ""

# Step 7: Enroll Bob
echo -e "${BLUE}Step 7: Enrolling Bob in courses${NC}"
./bin/gosql enrollment-create -u 2 -c 1 > /dev/null
./bin/gosql enrollment-create -u 2 -c 2 > /dev/null
echo -e "${GREEN}✓ Bob now enrolled in 2 courses${NC}"
echo ""

# Step 8: View Alice's enrollments
echo -e "${BLUE}Step 8: Alice's enrollments${NC}"
./bin/gosql enrollment-by-user -u 1
echo ""

# Step 9: View Go Basics enrollments
echo -e "${BLUE}Step 9: Students enrolled in Go Basics${NC}"
./bin/gosql enrollment-by-course -c 1
echo ""

# Step 10: Complete and cancel enrollments
echo -e "${BLUE}Step 10: Alice completes Go Basics${NC}"
./bin/gosql enrollment-complete -u 1 -c 1
echo ""

echo -e "${BLUE}Step 11: Bob cancels Python Advanced${NC}"
./bin/gosql enrollment-cancel -u 2 -c 2
echo ""

# Step 12: Final state
echo -e "${BLUE}Step 12: All enrollments (final state)${NC}"
./bin/gosql enrollment-list
echo ""

# Summary
echo -e "${GREEN}=== Transaction Benefits Demonstrated ===${NC}"
echo "✓ Atomicity: Enrollments only created when ALL checks pass"
echo "✓ Consistency: UNIQUE constraint prevents duplicate enrollments"
echo "✓ Foreign Keys: Can't enroll non-existent users or courses"
echo "✓ Data Integrity: Invalid operations rollback without side effects"
echo ""
echo -e "${BLUE}Key Points:${NC}"
echo "1. User/course existence verified in transaction"
echo "2. If ANY check fails, transaction rolls back"
echo "3. No partial state or orphaned records"
echo "4. UNIQUE constraint enforced at database level"
echo ""

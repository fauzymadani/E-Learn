#!/bin/bash

# Test PDF Upload Script
# This script tests file uploads to help debug why PDFs might not be uploading

BASE_URL="http://localhost:8080"
COURSE_ID=2

echo "==================================="
echo "PDF Upload Debug Test"
echo "==================================="
echo ""

# Check if server is running
if ! curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
    echo "❌ Server is not running on ${BASE_URL}"
    echo "Please start the server first: go run cmd/api/main.go"
    exit 1
fi

echo "✓ Server is running"
echo ""

# Get auth token
read -p "Enter your auth token (from login): " TOKEN
if [ -z "$TOKEN" ]; then
    echo "❌ Token is required"
    exit 1
fi
echo ""

# Create a test PDF file
TEST_PDF="test_upload.pdf"
if [ ! -f "$TEST_PDF" ]; then
    echo "Creating test PDF file..."
    # Create a simple PDF using echo and printf
    cat > "$TEST_PDF" << 'EOF'
%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Test PDF) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000317 00000 n
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
409
%%EOF
EOF
    echo "✓ Test PDF created: $TEST_PDF"
else
    echo "✓ Using existing test PDF: $TEST_PDF"
fi
echo ""

# Test 1: Upload only PDF
echo "==================================="
echo "Test 1: Upload PDF only"
echo "==================================="
echo "Request:"
echo "  URL: POST ${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons"
echo "  Fields: title, content, file (PDF)"
echo ""

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" \
    -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons" \
    -H "Authorization: Bearer ${TOKEN}" \
    -F "title=PDF Upload Test" \
    -F "content=Testing PDF upload functionality" \
    -F "file=@${TEST_PDF}")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE:/d')

echo "Response:"
echo "  HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" -eq 201 ]; then
    echo "  ✓ SUCCESS - Lesson created!"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"

    # Check if file_url is present
    FILE_URL=$(echo "$BODY" | jq -r '.file_url' 2>/dev/null)
    if [ "$FILE_URL" != "null" ] && [ ! -z "$FILE_URL" ]; then
        echo ""
        echo "  ✓ PDF URL: $FILE_URL"

        # Try to access the file
        echo "  Checking if file is accessible..."
        if curl -s -I "${BASE_URL}${FILE_URL}" | grep -q "200 OK"; then
            echo "  ✓ File is accessible at ${BASE_URL}${FILE_URL}"
        else
            echo "  ❌ File is NOT accessible"
        fi
    else
        echo "  ❌ file_url is empty or null!"
    fi
else
    echo "  ❌ FAILED"
    echo "$BODY"
fi
echo ""

# Test 2: Upload PDF with different extension case
echo "==================================="
echo "Test 2: Upload with .PDF (uppercase)"
echo "==================================="

# Copy test file with uppercase extension
cp "$TEST_PDF" "test_upload.PDF"

RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}\n" \
    -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons" \
    -H "Authorization: Bearer ${TOKEN}" \
    -F "title=PDF Upload Test (Uppercase)" \
    -F "content=Testing PDF upload with uppercase extension" \
    -F "file=@test_upload.PDF")

HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_CODE:/d')

echo "Response:"
echo "  HTTP Status: $HTTP_CODE"
if [ "$HTTP_CODE" -eq 201 ]; then
    echo "  ✓ SUCCESS - Case-insensitive check works!"
    FILE_URL=$(echo "$BODY" | jq -r '.file_url' 2>/dev/null)
    echo "  PDF URL: $FILE_URL"
else
    echo "  ❌ FAILED"
    echo "$BODY"
fi
echo ""

# Check server logs
echo "==================================="
echo "Check Server Logs"
echo "==================================="
echo "Look for these lines in your server output:"
echo "  - 'Received file field' messages"
echo "  - 'PDF uploaded successfully' messages"
echo "  - Any error messages about file uploads"
echo ""

# Check uploads directory
echo "==================================="
echo "Checking uploads directory"
echo "==================================="
cd "$(dirname "$0")"
if [ -d "uploads/files" ]; then
    FILE_COUNT=$(find uploads/files -type f | wc -l)
    echo "✓ uploads/files directory exists"
    echo "  Files in uploads/files: $FILE_COUNT"
    if [ $FILE_COUNT -gt 0 ]; then
        echo "  Latest files:"
        ls -lht uploads/files | head -5
    fi
else
    echo "❌ uploads/files directory does not exist!"
fi

if [ -d "uploads/videos" ]; then
    VIDEO_COUNT=$(find uploads/videos -type f | wc -l)
    echo "✓ uploads/videos directory exists"
    echo "  Files in uploads/videos: $VIDEO_COUNT"
else
    echo "❌ uploads/videos directory does not exist!"
fi
echo ""

# Cleanup
rm -f test_upload.PDF

echo "==================================="
echo "Troubleshooting Tips"
echo "==================================="
echo "If PDFs are not uploading:"
echo ""
echo "1. Check the server logs for:"
echo "   - 'Received file field \"file\"' - confirms file was received"
echo "   - 'PDF uploaded successfully' - confirms save worked"
echo "   - Error messages about permissions or disk space"
echo ""
echo "2. Check file permissions:"
echo "   ls -la uploads/"
echo "   ls -la uploads/files/"
echo ""
echo "3. Check disk space:"
echo "   df -h ."
echo ""
echo "4. Verify you're using the correct field name 'file' (not 'pdf')"
echo ""
echo "5. Make sure Content-Type header is NOT manually set"
echo "   (let curl/browser set it automatically for multipart)"
echo ""


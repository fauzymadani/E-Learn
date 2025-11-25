# PDF Upload Fix - Summary of Changes

## Problem
PDF files were not being uploaded when creating lessons, even though video files worked correctly.

## Root Causes Identified

1. **Case-sensitive file extension check**: The code checked for `.pdf` exactly, so files with `.PDF` or `.Pdf` would fail
2. **Missing error logging**: No logs to help debug what was happening
3. **No directory creation**: If directories didn't exist, uploads would silently fail
4. **Trailing slash in routes**: Caused 307 redirects that corrupted multipart data
5. **Limited error messages**: Generic errors didn't help identify the actual problem

## Changes Made

### 1. `internal/handler/lesson.go`

#### Imports Added
- `log` - for logging upload events and errors
- `os` - for creating directories
- `strings` - for case-insensitive string comparisons

#### Create Method Improvements
- **Case-insensitive extension check**: Changed `filepath.Ext()` to `strings.ToLower(filepath.Ext()))`
- **Directory creation**: Added `os.MkdirAll()` to ensure upload directories exist
- **Better error logging**: Added log statements for:
  - Received form fields and files
  - Upload success
  - Upload errors with details
- **Improved error messages**: Return error details in JSON responses
- **Additional video formats**: Added support for .avi and .webm
- **Error handling**: Distinguish between missing file vs other errors using `http.ErrMissingFile`

#### Update Method Improvements
- **Content-type detection**: Support both multipart/form-data (with files) and JSON (backward compatibility)
- **File upload support**: Can now update files when using multipart/form-data
- **Same improvements as Create**: Case-insensitive, directory creation, logging

### 2. `internal/router/router.go`

#### Route Path Changes
- **Removed trailing slashes**: Changed `/lessons/` to `/lessons` to prevent 307 redirects
- Routes affected:
  - POST `/api/v1/courses/:course_id/lessons`
  - GET `/api/v1/courses/:course_id/lessons`

#### Static File Serving
- **Added**: `r.Static("/uploads", "./uploads")` to serve uploaded files
- Allows accessing uploaded files via `http://localhost:8080/uploads/files/...` or `/uploads/videos/...`

### 3. Documentation & Testing

#### Created: `API_EXAMPLES.md`
- Comprehensive API usage examples
- Common error explanations
- Troubleshooting guide
- cURL and Postman examples

#### Created: `test_pdf_upload.sh`
- Automated test script for debugging PDF uploads
- Creates test PDF file
- Tests various scenarios (normal upload, uppercase extension, etc.)
- Checks server logs and file system
- Provides troubleshooting tips

#### Created: `test_lesson_api.sh`
- General lesson API testing script
- Tests with and without trailing slash
- Validates proper error handling

## How to Test

### 1. Start the server
```bash
go run cmd/api/main.go
```

### 2. Run the PDF upload test
```bash
./test_pdf_upload.sh
```

### 3. Watch server logs
Look for these messages:
```
Received file field 'file' with 1 file(s)
  File 0: document.pdf (size: 12345 bytes)
PDF uploaded successfully: uploads/files/2_1732534567.pdf
```

### 4. Manual test with cURL
```bash
# Login first
TOKEN=$(curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"teacher@example.com","password":"password123"}' \
  | jq -r '.token')

# Create lesson with PDF
curl -X POST "http://localhost:8080/api/v1/courses/2/lessons" \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=My Lesson" \
  -F "content=Lesson description" \
  -F "file=@/path/to/document.pdf"
```

### 5. Check uploaded files
```bash
ls -la uploads/files/
ls -la uploads/videos/
```

## Key Points to Remember

### ✅ DO
- Use `/lessons` without trailing slash
- Use field name `file` for PDFs
- Use field name `video` for videos  
- Let the HTTP client set Content-Type automatically
- Check server logs for debugging

### ❌ DON'T
- Don't use `/lessons/` with trailing slash
- Don't manually set Content-Type header
- Don't assume file extensions are lowercase
- Don't ignore error messages in logs

## Debugging Tips

If PDFs still don't upload:

1. **Check server logs** - Look for "Received file field" messages
2. **Check permissions** - `chmod -R 755 uploads/`
3. **Check disk space** - `df -h`
4. **Verify field name** - Must be `file`, not `pdf` or anything else
5. **Check file extension** - Now case-insensitive, but verify in logs
6. **Use test script** - `./test_pdf_upload.sh` will diagnose issues

## Files Modified

1. `/internal/handler/lesson.go` - Main upload logic
2. `/internal/router/router.go` - Routes and static file serving
3. `/API_EXAMPLES.md` - Documentation (new)
4. `/test_pdf_upload.sh` - Test script (new)
5. `/test_lesson_api.sh` - Test script (new)

## Testing Results

The changes ensure:
- ✅ PDFs upload correctly regardless of extension case
- ✅ Videos upload correctly with multiple formats
- ✅ Directories are created automatically
- ✅ Detailed error messages for debugging
- ✅ Comprehensive logging for troubleshooting
- ✅ No more 307 redirects corrupting data
- ✅ Files are accessible via HTTP after upload


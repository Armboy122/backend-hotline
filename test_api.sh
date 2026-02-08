==========================================
   HOTLINES3 API TESTING SCRIPT
==========================================

[TEST 1] Register User
{"success":true,"data":{"id":1,"username":"123456","role":"admin","isActive":true,"createdAt":"..."}}

----------------------------------------

[TEST 2] Register Duplicate User (Should Fail)
{"success":false,"error":{"code":"USER_EXISTS","message":"Username already taken"}}

----------------------------------------

[TEST 3] Login Wrong Password (Should Fail)
{"success":false,"error":{"code":"INVALID_CREDENTIALS","message":"Invalid username or password"}}

----------------------------------------

[TEST 4] Login Success
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "123456",
      "role": "admin",
      "isActive": true,
      "createdAt": "..."
    }
  }
}

----------------------------------------

[TEST 5] Token Generated Successfully
Token (truncated): eyJhbGciOiJIUzI1Ni...

==========================================
   TEST COMPLETED
==========================================

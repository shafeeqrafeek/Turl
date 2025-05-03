# Turl - URL Shortener

### Approach

I am beginning this project by using `net/http` package for creating the API. 
The idea is to create a basic `Server` and add `Handlers` for each endpoint. 

**TODO**: Upgrade the API to a gin framework. 

### Sample Spec

1. POST /shorten

Request Body (JSON):

````
{
  "url": "https://example.com/very/long/link"
}
````
Response:

````
{
  "short_url": "http://localhost:8080/t/abc123"
}
````

2. GET /urls

Response:

````
[
  {
    "short_url": "http://localhost:8080/t/abc123",
    "original_url": "https://example.com/very/long/link"
  },
  {
    "short_url": "http://localhost:8080/t/xyz456",
    "original_url": "https://openai.com/research"
  }
]

````
3. GET /t/abc123

````
Redirect (HTTP 302) to:
https://example.com/very/long/link
````


# Implementation

1. how to read the body of an incoming request?

**previous approach:**

make a byte array with the size of the incoming request's body. 
and manually read the request body into the byte array and then unmarshall it into a struct 

**improved approach:** 

create a new json decoder for the request body and decode it directly into the struct 

2. how to define request methods for a particular route?

By defining the method along with url pattern. e.g., "GET /urls"

# Extensively used Packages

1. net/http
2. net/url - tiny package helpful for url parsing related stuff
3. encoding/json

For testing:

1. testing
2. github.com/stretchr/testify (assert and require)

# Code Optimization takeaways

1. dont nest error handling if conditions inside error handling if conditions. write separate functions for each error 
handling. like WriteError function which is used inside a *if err != nil* block.  

# Issues to Fix

1. how to create live server and auto-detect changes?
2. Input validation for URL - using net/url package
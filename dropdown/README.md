# Dropdown Service

This service provides dropdown data for various form fields used throughout the application.

## Endpoints

### GET /v1/dropdown/service-type
Returns available service types.

**Response:**
```json
{
  "code": 200,
  "data": [
    {"value": "cargo", "text": "Cargo"},
    {"value": "transit", "text": "Transit"}
  ],
  "message": "Service types retrieved successfully"
}
```

### GET /v1/dropdown/shipping-type
Returns available shipping types.

**Response:**
```json
{
  "code": 200,
  "data": [
    {"value": "sea", "text": "Sea"},
    {"value": "air", "text": "Air"}
  ],
  "message": "Shipping types retrieved successfully"
}
```

## Usage

These endpoints can be used to populate dropdown menus in forms across the application. The data is currently static but the service is designed to be easily extended to support database-driven dropdown data in the future.

## Authentication

Both endpoints require JWT authentication via the Authorization header.
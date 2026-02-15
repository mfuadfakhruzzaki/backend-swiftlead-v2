import json
import uuid
import os

# Collection Info
COLLECTION_NAME = "SwiftLead v2 API"
BASE_URL = "{{base_url}}"

# Helper to create a unique ID
def get_id():
    return str(uuid.uuid4())

# Helper to create a folder
def create_folder(name, item=None):
    return {
        "name": name,
        "item": item if item else [],
        "description": ""
    }

# Helper to create a request
def create_request(name, method, url_path, body=None, query_params=None, auth_type="bearer", description=""):
    request = {
        "method": method,
        "header": [
            {
                "key": "Content-Type",
                "value": "application/json",
                "type": "text"
            },
            {
                "key": "Accept",
                "value": "application/json",
                "type": "text"
            }
        ],
        "url": {
            "raw": f"{BASE_URL}{url_path}",
            "host": ["{{base_url}}"],
            "path": url_path.strip("/").split("/"),
            "query": query_params if query_params else []
        },
        "description": description
    }

    if body:
        request["body"] = {
            "mode": "raw",
            "raw": json.dumps(body, indent=2)
        }

    item = {
        "name": name,
        "request": request,
        "response": []
    }
    
    return item

# Define specific folder structures
folders = []

# Auth Folder
auth_requests = [
    create_request("Register (Public)", "POST", "/api/v1/auth/register", {
        "name": "John Doe",
        "email": "john@example.com",
        "password": "SecurePass123!@",
        "phone": "081234567890"
    }, description="Register a new user"),
    create_request("Login", "POST", "/api/v1/auth/login", {
        "email": "john@example.com",
        "password": "SecurePass123!@"
    }, description="Login to get token"),
    create_request("Change Password", "POST", "/api/v1/auth/change-password", {
        "old_password": "OldPass123!@",
        "new_password": "NewPass456!@"
    }, description="Change current user password"),
    create_request("Forgot Password (Admin)", "POST", "/api/v1/auth/forgot-password", {
        "email": "user@example.com"
    }, description="Admin reset user password"),
    create_request("Admin Register User", "POST", "/api/v1/auth/admin/register", {
        "name": "New Technician",
        "email": "tech@example.com",
        "password": "TechPass123!@",
        "role": "technician",
        "phone": "081234567890"
    }, description="Admin create new user")
]
folders.append(create_folder("Auth", auth_requests))

# User Folder
user_requests = [
    create_request("Get My Profile", "GET", "/api/v1/users/me"),
    create_request("Update My Profile", "PATCH", "/api/v1/users/me", {
        "name": "Updated Name",
        "phone": "089876543210",
        "avatar_url": "https://example.com/avatar.jpg"
    }),
    create_request("List Users (Admin)", "GET", "/api/v1/users", query_params=[
        {"key": "page", "value": "1"},
        {"key": "limit", "value": "20"},
        {"key": "role", "value": "farmer"}
    ]),
    create_request("Create User (Admin)", "POST", "/api/v1/users", {
        "name": "New Technician",
        "email": "tech@example.com",
        "password": "TechPass123!@",
        "role": "technician",
        "phone": "081234567890"
    })
]
folders.append(create_folder("Users", user_requests))

# RBW Folder
rbw_requests = [
    create_request("Create RBW", "POST", "/api/v1/rbw", {
        "code": "RBW-001",
        "name": "Rumah Walet Jakarta",
        "address": "Jl. Merdeka No. 1",
        "latitude": -6.2088,
        "longitude": 106.8456,
        "total_floors": 3,
        "description": "Walet house description"
    }),
    create_request("List RBWs", "GET", "/api/v1/rbw", query_params=[
        {"key": "page", "value": "1"},
        {"key": "limit", "value": "20"}
    ]),
    create_request("Get RBW", "GET", "/api/v1/rbw/{{rbw_id}}"),
    create_request("Update RBW", "PATCH", "/api/v1/rbw/{{rbw_id}}", {
        "name": "Updated Name",
        "total_floors": 4
    }),
    create_request("Delete RBW", "DELETE", "/api/v1/rbw/{{rbw_id}}"),
    create_request("List RBW Nodes", "GET", "/api/v1/rbw/{{rbw_id}}/nodes"),
    create_request("Create RBW Node", "POST", "/api/v1/rbw/{{rbw_id}}/nodes", {
        "node_type": "gateway",
        "node_code": "GW-001",
        "esp32_uid": "AA:BB:CC:DD:EE:FF",
        "has_audio": True,
        "has_pump": True
    }),
    create_request("List RBW Alerts", "GET", "/api/v1/rbw/{{rbw_id}}/alerts"),
    create_request("List RBW Harvests", "GET", "/api/v1/rbw/{{rbw_id}}/harvests"),
    create_request("List RBW Transactions", "GET", "/api/v1/rbw/{{rbw_id}}/transactions")
]
folders.append(create_folder("RBW", rbw_requests))

# Node Folder
node_requests = [
    create_request("Get Node", "GET", "/api/v1/nodes/{{node_id}}"),
    create_request("Update Node", "PATCH", "/api/v1/nodes/{{node_id}}", {
        "node_code": "GW-UPDATED"
    }),
    create_request("Delete Node", "DELETE", "/api/v1/nodes/{{node_id}}"),
    create_request("List Node Sensors", "GET", "/api/v1/nodes/{{node_id}}/sensors"),
    create_request("Create Node Sensor", "POST", "/api/v1/nodes/{{node_id}}/sensors", {
        "sensor_type": "temp",
        "sensor_name": "Temperature Sensor 1",
        "unit": "°C"
    }),
    create_request("Get Audio State", "GET", "/api/v1/nodes/{{node_id}}/audio"),
    create_request("Control Audio", "PATCH", "/api/v1/nodes/{{node_id}}/audio", {
        "action": "audio_set_lmb",
        "value": 1
    }),
    create_request("Control Pump", "PATCH", "/api/v1/nodes/{{node_id}}/pump", {
        "action": "sprayer_set",
        "value": 1
    })
]
folders.append(create_folder("Nodes", node_requests))

# Sensor Folder
sensor_requests = [
    create_request("Get Sensor", "GET", "/api/v1/sensors/{{sensor_id}}"),
    create_request("Update Sensor", "PATCH", "/api/v1/sensors/{{sensor_id}}", {
        "sensor_name": "Updated Sensor Name"
    }),
    create_request("Create Sensor Reading", "POST", "/api/v1/sensors/{{sensor_id}}/readings", {
        "value": 28.5
    }),
    create_request("Get Sensor Readings", "GET", "/api/v1/sensors/{{sensor_id}}/readings", query_params=[
        {"key": "limit", "value": "100"}
    ]),
    create_request("Get Sensor Trend", "GET", "/api/v1/sensors/{{sensor_id}}/trend")
]
folders.append(create_folder("Sensors", sensor_requests))

# Alert Folder
alert_requests = [
    create_request("List All Alerts", "GET", "/api/v1/alerts", query_params=[
        {"key": "page", "value": "1"},
        {"key": "limit", "value": "20"},
        {"key": "is_read", "value": "false"}
    ]),
    create_request("Mark Alert Read", "PATCH", "/api/v1/alerts/{{alert_id}}/read"),
    create_request("Resolve Alert", "PATCH", "/api/v1/alerts/{{alert_id}}/resolve")
]
folders.append(create_folder("Alerts", alert_requests))

# Harvest Folder
harvest_requests = [
    create_request("Create Harvest", "POST", "/api/v1/harvests", {
        "rbw_id": "{{rbw_id}}",
        "floor_no": 1,
        "harvested_at": "2026-02-09T08:00:00Z",
        "nests_count": 50,
        "weight_kg": 2.5,
        "grade": "good",
        "notes": "Test harvest"
    }),
    create_request("List Harvests", "GET", "/api/v1/harvests", query_params=[
        {"key": "page", "value": "1"},
        {"key": "limit", "value": "20"}
    ]),
    create_request("Get Harvest", "GET", "/api/v1/harvests/{{harvest_id}}"),
    create_request("Update Harvest", "PATCH", "/api/v1/harvests/{{harvest_id}}", {
        "weight_kg": 2.8,
        "notes": "Updated notes"
    }),
    create_request("Delete Harvest", "DELETE", "/api/v1/harvests/{{harvest_id}}"),
    create_request("Get Harvest Stats", "GET", "/api/v1/harvests/stats", query_params=[
        {"key": "rbw_id", "value": "{{rbw_id}}"}
    ])
]
folders.append(create_folder("Harvests", harvest_requests))

# Service Request Folder
service_requests = [
    create_request("Create Service Request", "POST", "/api/v1/service-requests", {
        "rbw_id": "{{rbw_id}}",
        "type": "maintenance",
        "issue": "Sensor malfunction"
    }),
    create_request("List Service Requests", "GET", "/api/v1/service-requests", query_params=[
        {"key": "status", "value": "pending"}
    ]),
    create_request("Get Service Request", "GET", "/api/v1/service-requests/{{service_request_id}}"),
    create_request("Update Service Request", "PATCH", "/api/v1/service-requests/{{service_request_id}}", {
        "status": "in_progress",
        "assigned_to": "{{user_id}}"
    })
]
folders.append(create_folder("Service Requests", service_requests))

# Transaction Folder
transaction_requests = [
    create_request("Create Transaction", "POST", "/api/v1/transactions", {
        "rbw_id": "{{rbw_id}}",
        "category_id": "{{category_id}}",
        "amount": 500000,
        "type": "income",
        "transaction_date": "2026-02-09T00:00:00Z"
    }),
    create_request("Get Transaction", "GET", "/api/v1/transactions/{{transaction_id}}"),
    create_request("Update Transaction", "PATCH", "/api/v1/transactions/{{transaction_id}}", {
        "amount": 600000
    }),
    create_request("Delete Transaction", "DELETE", "/api/v1/transactions/{{transaction_id}}")
]
folders.append(create_folder("Transactions", transaction_requests))

# Category Folder
category_requests = [
    create_request("List Categories", "GET", "/api/v1/transaction-categories"),
    create_request("Create Category", "POST", "/api/v1/transaction-categories", {
        "name": "New Category",
        "type": "income",
        "description": "Category description"
    }),
    create_request("Update Category", "PATCH", "/api/v1/transaction-categories/{{category_id}}", {
        "name": "Updated Category"
    }),
    create_request("Delete Category", "DELETE", "/api/v1/transaction-categories/{{category_id}}")
]
folders.append(create_folder("Transaction Categories", category_requests))

# Financial Statement Folder
financial_requests = [
    create_request("Generate Statement", "POST", "/api/v1/financial-statements", {
        "rbw_id": "{{rbw_id}}",
        "start_date": "2026-01-01",
        "end_date": "2026-12-31"
    })
]
folders.append(create_folder("Financial Statements", financial_requests))

# AI Folder
ai_requests = [
    create_request("Health Check", "GET", "/api/v1/ai/health"),
    create_request("Predict Grade", "POST", "/api/v1/ai/predict-grade", {
        "temperature": 28.5,
        "humidity": 75.0,
        "ammonia": 15.0
    }),
    create_request("Predict Pump", "POST", "/api/v1/ai/predict-pump", {
        "temperature": 32.0,
        "humidity": 65.0,
        "ammonia": 20.0
    }),
    create_request("Analyze", "POST", "/api/v1/ai/analyze", {
        "temperature": 28.5,
        "humidity": 75.0,
        "ammonia": 15.0
    }),
    create_request("Anomaly Detect", "POST", "/api/v1/ai/anomaly-detect", {
        "sensor_type": "temp",
        "value": 45.0
    })
]
folders.append(create_folder("AI Engine", ai_requests))

# Upload Folder
upload_requests = [
    create_request("Upload Avatar", "POST", "/api/v1/uploads/avatar", description="Requires multipart/form-data with 'file' field"),
    create_request("Upload RBW Photo", "POST", "/api/v1/uploads/rbw/{{rbw_id}}/photo", description="Requires multipart/form-data with 'file' field")
]
folders.append(create_folder("Uploads", upload_requests))

# WebSocket
ws_requests = [
    create_request("WebSocket Connection", "GET", "/api/v1/ws", query_params=[{"key": "token", "value": "{{token}}"}], description="WebSocket connection URL"),
    create_request("WebSocket Stats", "GET", "/api/v1/ws/stats")
]
folders.append(create_folder("WebSocket", ws_requests))

# Health
health_requests = [
    create_request("Health Check", "GET", "/health"),
    create_request("Prometheus Metrics", "GET", "/metrics")
]
folders.append(create_folder("Health & Metrics", health_requests))


# Construct full collection
collection = {
    "info": {
        "name": COLLECTION_NAME,
        "_postman_id": get_id(),
        "description": "API Documentation for SwiftLead v2",
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    },
    "item": folders,
    "auth": {
        "type": "bearer",
        "bearer": [
            {
                "key": "token",
                "value": "{{token}}",
                "type": "string"
            }
        ]
    }
}

# Write to file
with open("SwiftLead_v2.postman_collection.json", "w") as f:
    json.dump(collection, f, indent=4)

print("Collection generated successfully: SwiftLead_v2.postman_collection.json")

# Environment file
environment = {
    "name": "SwiftLead v2 Env",
    "values": [
        {"key": "base_url", "value": "https://api.swiftlead.fuadfakhruz.com", "enabled": True},
        {"key": "token", "value": "", "enabled": True},
        {"key": "rbw_id", "value": "", "enabled": True},
        {"key": "node_id", "value": "", "enabled": True},
        {"key": "sensor_id", "value": "", "enabled": True},
        {"key": "alert_id", "value": "", "enabled": True},
        {"key": "harvest_id", "value": "", "enabled": True},
        {"key": "service_request_id", "value": "", "enabled": True},
        {"key": "transaction_id", "value": "", "enabled": True},
        {"key": "category_id", "value": "", "enabled": True},
        {"key": "user_id", "value": "", "enabled": True}
    ]
}

with open("SwiftLead_v2.postman_environment.json", "w") as f:
    json.dump(environment, f, indent=4)

print("Environment generated successfully: SwiftLead_v2.postman_environment.json")

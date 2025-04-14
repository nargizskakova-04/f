☕ Coffee Boom

Coffee Boom is a backend service for a fast-paced restaurant or café, designed to efficiently manage orders, menu items, inventory, and analytics — all powered by PostgreSQL and built with Go.

From order processing to detailed sales reports and inventory tracking, Coffee Boom helps businesses stay organized, informed, and scalable.
🌐 Overview

This system provides a RESTful API for managing a restaurant's day-to-day operations:

    Accept and manage customer orders

    Track order statuses and history

    Organize menu items and ingredients

    Maintain real-time inventory levels

    View detailed sales reports and search insights

    Automatically track pricing and inventory changes

Built with a focus on performance, reliability, and scalability.
🧰 Tech Stack

    Backend Language: Go (Golang)

    Database: PostgreSQL

    Containerization: Docker & Docker Compose

    Architecture: Layered (Handlers → Services → Repositories)

🚀 Getting Started
1. Clone the Repository

git clone https://github.com/your-username/coffee-boom.git
cd coffee-boom

2. Start the Application

docker compose up

The API will be available at:
📍 http://localhost:8080
🗃️ Key Features
✅ Orders

    Place new orders

    Modify or cancel existing orders

    Track order status and history

    Close & archive completed orders

🍽️ Menu Management

    Add or remove menu items

    Define ingredients, categories, and prices

    Monitor changes in price over time

📦 Inventory Management

    Track current stock of ingredients

    Record inventory usage automatically on orders

    Generate reports on leftovers and transactions

📊 Analytics & Reports

    View sales totals

    Analyze popular items

    Filter data by period or category

    Search orders and menu using full-text queries

📑 API Endpoints
Orders

POST    /orders
GET     /orders
GET     /orders/{id}
PUT     /orders/{id}
DELETE  /orders/{id}
POST    /orders/{id}/close

Menu Items

POST    /menu
GET     /menu
GET     /menu/{id}
PUT     /menu/{id}
DELETE  /menu/{id}

Inventory

POST    /inventory
GET     /inventory
GET     /inventory/{id}
PUT     /inventory/{id}
DELETE  /inventory/{id}

Reports

GET /reports/total-sales
GET /reports/popular-items
GET /orders/numberOfOrderedItems?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD
GET /reports/search?q=espresso&filter=menu,orders&minPrice=1000
GET /reports/orderedItemsByPeriod?period=month&year=2025
GET /inventory/getLeftOvers?sortBy=quantity&page=1&pageSize=10

🧱 Database Structure

The database schema is designed with normalization and scalability in mind:

    Relational tables with foreign keys and constraints

    Use of ENUM for predefined status/roles

    JSONB for flexible fields (like special instructions)

    ARRAY types for tags or labels

    Historical tables for prices, status changes, and inventory logs

All schema definitions and test data are initialized via init.sql.
📦 Deployment

Coffee Boom is fully containerized and ready to be deployed on any server or cloud provider supporting Docker.

docker compose up -d

Environment variables for the database can be configured in .env or passed directly into the docker-compose.yml file.
🧪 Sample Data

Out of the box, the service starts with:

    ✅ 30+ orders

    ✅ 10+ menu items

    ✅ Full inventory setup

    ✅ Sample pricing & order history

    ✅ Full-text searchable content
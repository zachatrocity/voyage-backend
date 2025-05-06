<div align="center">
  
# ✉️ Voyage ✈️


A (WIP) notmuch travel plan aggregator that searches through your email to identify and organize travel-related information.

</div>

## Overview

Voyage is designed to simplify travel planning by automatically aggregating confirmation emails and itineraries from your inbox. It creates a consolidated view of your travel plans, making it easy to keep track of flights, accommodations, car rentals, and other travel arrangements.

## Features

- **Email Integration**: Automatically scans and processes travel-related emails
- **Trip Organization**: Groups related travel items into coherent trips
- **Self-Hosted**: Full control over your data with Docker-based deployment

## Architecture

Voyage consists of several key components:

1. **Email Fetching Service**: Uses mbsync to retrieve emails from your accounts
2. **Notmuch Integration**: Indexes and tags emails for efficient searching
3. **REST API**: Simple Go API for searching the notmuch database

## API Usage

Search the notmuch database:
```
GET /api/v1/search?q=airbnb
GET /api/v1/emails/{message_id}
```
